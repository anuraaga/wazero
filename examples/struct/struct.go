package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/wasi"
	"log"
	"unsafe"
)

// structWasm was compiled using `tinygo build -o struct.wasm -scheduler=none --no-debug -target=wasi struct.go`
//go:embed testdata/struct.wasm
var structWasm []byte

type bear struct {
	name     string
	greeting string
}

// See README.md for a full description.
func main() {
	// Choose the context to use for function calls.
	ctx := context.Background()

	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntime()

	// Instantiate a module named "env" that exports operations on a bear.
	env, err := r.NewModuleBuilder("env").
		ExportFunction("setName", setName).
		ExportFunction("getName", getName).
		Instantiate(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer env.Close()

	// Note: testdata/greet.go doesn't use WASI, but TinyGo needs it to
	// implement functions such as panic.
	wm, err := wasi.InstantiateSnapshotPreview1(ctx, r)
	if err != nil {
		log.Fatal(err)
	}
	defer wm.Close()

	// Instantiate a module that imports the "getName" and "setName" functions defined in "env".
	mod, err := r.InstantiateModuleFromCode(ctx, structWasm)
	if err != nil {
		log.Fatal(err)
	}
	defer mod.Close()

	// Get a references to functions we'll use in this example.
	hello := mod.ExportedFunction("hello")

	b := &bear{
		name:     "Yogi",
		greeting: "Yabba Dabba Doo!",
	}

	_, err = hello.Call(ctx, b.ptr())
	if err != nil {
		log.Fatal(err)
	}

	// Name should have been updated in wasm
	fmt.Println("name:" + b.name)
	// Make sure other fields are still there too
	fmt.Println("greeting:" + b.greeting)
}

func setName(m api.Module, bearPtr uint64, nameOff uint32, nameLen uint32) {
	buf, ok := m.Memory().Read(nameOff, nameLen)
	if !ok {
		log.Fatalf("Memory.Read(%d, %d) out of range", nameOff, nameLen)
	}
	b := bearFromPtr(bearPtr)
	b.name = string(buf)
}

func getName(m api.Module, bearPtr uint64, bufOff uint32) uint32 {
	b := bearFromPtr(bearPtr)

	if !m.Memory().Write(bufOff, []byte(b.name)) {
		log.Fatalf("Memory.Write(%d, %d) out of range of memory size %d",
			bufOff, len(b.name), m.Memory().Size())
	}

	return uint32(len(b.name))
}

func bearFromPtr(bearPtr uint64) *bear {
	return (*bear)(unsafe.Pointer(uintptr(bearPtr)))
}

func (b *bear) ptr() uint64 {
	return uint64(uintptr(unsafe.Pointer(b)))
}
