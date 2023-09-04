package experimental_test

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"runtime"
	"sync/atomic"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// pthreadWasm was generated by the following:
//
//	docker run -it --rm -v `pwd`/testdata:/workspace ghcr.io/webassembly/wasi-sdk:wasi-sdk-20 sh -c '$CC -o /workspace/pthread.wasm /workspace/pthread.c --target=wasm32-wasi-threads --sysroot=/wasi-sysroot -pthread -mexec-model=reactor -Wl,--export=run -Wl,--export=get'
//
// TODO: Use zig cc instead of wasi-sdk to compile when it supports wasm32-wasi-threads
// https://github.com/ziglang/zig/issues/15484
//
//go:embed testdata/pthread.wasm
var pthreadWasm []byte

//go:embed testdata/memory.wasm
var memoryWasm []byte

// This shows how to use a WebAssembly module compiled with the threads feature.
func ExampleCoreFeaturesThreads() {
	// Use a default context
	ctx := context.Background()

	// Threads support must be enabled explicitly in addition to standard V2 features.
	cfg := wazero.NewRuntimeConfig().WithCoreFeatures(api.CoreFeaturesV2 | experimental.CoreFeaturesThreads)

	r := wazero.NewRuntimeWithConfig(ctx, cfg)
	defer r.Close(ctx)

	wasmCompiled, err := r.CompileModule(ctx, pthreadWasm)
	if err != nil {
		log.Panicln(err)
	}

	// Because we are using wasi-sdk to compile the guest, we must initialize WASI.
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	if _, err := r.InstantiateWithConfig(ctx, memoryWasm, wazero.NewModuleConfig().WithName("env")); err != nil {
		log.Panicln(err)
	}

	mod, err := r.InstantiateModule(ctx, wasmCompiled, wazero.NewModuleConfig().WithStartFunctions("_initialize"))
	if err != nil {
		log.Panicln(err)
	}

	// Channel to synchronize start of goroutines before running.
	startCh := make(chan struct{})
	// Channel to synchronize end of goroutines.
	endCh := make(chan struct{})

	// We start up 8 goroutines and run for 6000 iterations each. The count should reach
	// 48000, at the end, but it would not if threads weren't working!
	for i := 0; i < 8; i++ {
		go func() {
			defer func() { endCh <- struct{}{} }()
			<-startCh

			// We must instantiate a child per simultaneous thread. This should normally be pooled
			// among arbitrary goroutine invocations.
			child := createChildModule(r, mod, wasmCompiled)
			fn := child.mod.ExportedFunction("run")
			for i := 0; i < 6000; i++ {
				_, err := fn.Call(ctx)
				if err != nil {
					log.Panicln(err)
				}
			}
			runtime.KeepAlive(child)
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

type childModule struct {
	mod        api.Module
	tlsBasePtr uint32
	exitCh     chan struct{}
}

var prevTID uint32

func createChildModule(rt wazero.Runtime, root api.Module, wasmCompiled wazero.CompiledModule) *childModule {
	ctx := context.Background()

	// Not executing function so is at end of stack
	stackPointer := root.ExportedGlobal("__stack_pointer").Get()
	tlsBase := root.ExportedGlobal("__tls_base").Get()

	// Thread-local-storage for the main thread is from __tls_base to __stack_pointer
	size := stackPointer - tlsBase

	malloc := root.ExportedFunction("malloc")

	// Allocate memory for the child thread stack
	res, err := malloc.Call(ctx, size)
	if err != nil {
		panic(err)
	}
	ptr := uint32(res[0])

	child, err := rt.InstantiateModule(ctx, wasmCompiled, wazero.NewModuleConfig().
		// Don't need to execute start functions again in child, it crashes anyways because
		// LLVM only allows calling them once.
		WithStartFunctions())
	if err != nil {
		panic(err)
	}
	initTLS := child.ExportedFunction("__wasm_init_tls")
	if _, err := initTLS.Call(ctx, uint64(ptr)); err != nil {
		panic(err)
	}

	tid := atomic.AddUint32(&prevTID, 1)
	root.Memory().WriteUint32Le(ptr, ptr)
	root.Memory().WriteUint32Le(ptr+20, tid)
	child.ExportedGlobal("__stack_pointer").(api.MutableGlobal).Set(uint64(ptr) + size)

	ret := &childModule{
		mod:        child,
		tlsBasePtr: ptr,
	}
	runtime.SetFinalizer(ret, func(obj interface{}) {
		cm := obj.(*childModule)
		free := cm.mod.ExportedFunction("free")
		// Ignore errors since runtime may have been closed before this is called.
		_, _ = free.Call(ctx, uint64(cm.tlsBasePtr))
		_ = cm.mod.Close(context.Background())
	})
	return ret
}
