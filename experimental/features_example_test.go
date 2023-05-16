package experimental_test

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// pthreadWasm was generated by the following:
//
//	docker run -it --rm -v `pwd`/testdata:/workspace ghcr.io/webassembly/wasi-sdk:wasi-sdk-20 sh -c '$CC -o /workspace/pthread.wasm /workspace/pthread.c --target=wasm32-wasi-threads --sysroot=/wasi-sysroot -pthread -mexec-model=reactor -Wl,--export=run -Wl,--export=get'
//
//go:embed testdata/pthread.wasm
var pthreadWasm []byte

// This shows how to use a WebAssembly module compiled with the threads feature.
func ExampleCoreFeaturesThreads() {
	// Use a default context
	ctx := context.Background()

	// Threads support is currently only supported with interpreter, so the config
	// must explicitly specify it.
	cfg := wazero.NewRuntimeConfigInterpreter()

	// Threads support must be enabled explicitly in addition to standard V2 features.
	cfg = cfg.WithCoreFeatures(api.CoreFeaturesV2 | experimental.CoreFeaturesThreads)

	r := wazero.NewRuntimeWithConfig(ctx, cfg)
	defer r.Close(ctx)

	// Because we are using wasi-sdk to compile the guest, we must initialize WASI.
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	mod, err := r.Instantiate(ctx, pthreadWasm)
	if err != nil {
		log.Panicln(err)
	}

	// Channel to synchronize start of goroutines before running.
	startCh := make(chan struct{})
	// Channel to synchronize end of goroutines.
	endCh := make(chan struct{})

	// Unfortunately, while memory accesses are thread-safe using atomic operations, compilers such
	// as LLVM still have global state that is not handled thread-safe, preventing actually making
	// concurrent invocations. We go ahead and add a global lock for now until this is resolved.
	// TODO: Remove this lock once functions can actually be called concurrently.
	var mu sync.Mutex

	// We start up 8 goroutines and run for 6000 iterations each. The count should reach
	// 48000, at the end, but it would not if threads weren't working!
	for i := 0; i < 8; i++ {
		go func() {
			defer func() { endCh <- struct{}{} }()
			<-startCh

			mu.Lock()
			defer mu.Unlock()

			// ExportedFunction must be called within each goroutine to have independent call stacks.
			// This incurs some overhead, a sync.Pool can be used to reduce this overhead if neeeded.
			fn := mod.ExportedFunction("run")
			for i := 0; i < 6000; i++ {
				_, err := fn.Call(ctx)
				if err != nil {
					log.Panicln(err)
				}
			}
		}()
	}
	for i := 0; i < 8; i++ {
		startCh <- struct{}{}
	}
	for i := 0; i < 8; i++ {
		<-endCh
	}

	res, err := mod.ExportedFunction("get").Call(ctx)
	if err != nil {
		log.Panicln(err)
	}
	fmt.Println(res[0])
	// Output: 48000
}
