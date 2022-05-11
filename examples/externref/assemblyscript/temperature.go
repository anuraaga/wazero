package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/wasi"
	"log"
	"os"
	"strconv"
	"unsafe"
)

// temperatureWasm was compiled using `npm install && npm run asbuild:release`
//go:embed testdata/temperature.wasm
var temperatureWasm []byte

func main() {
	// Choose the context to use for function calls.
	ctx := context.Background()

	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntimeWithConfig(wazero.NewRuntimeConfigJIT().WithWasmCore2())
	defer r.Close(ctx) // This closes everything this Runtime created.

	// Instantiate a Go-defined module named "env" that exports a function to
	// log to the console.
	_, err := r.NewModuleBuilder("lib").
		ExportFunction("Calculator_Push", _calculator_Push).
		ExportFunction("Calculator_Pop", _calculator_Pop).
		ExportFunction("Calculator_Add", _calculator_Add).
		ExportFunction("Calculator_Subtract", _calculator_Subtract).
		ExportFunction("Calculator_Multiply", _calculator_Multiply).
		ExportFunction("Calculator_Divide", _calculator_Divide).
		Instantiate(ctx)
	if err != nil {
		log.Panicln(err)
	}

	// Note: testdata/temperature doesn't use WASI, but AssemblyScript needs it to
	// implement functions such as abort.
	if _, err = wasi.InstantiateSnapshotPreview1(ctx, r); err != nil {
		log.Panicln(err)
	}

	// Instantiate a WebAssembly module that imports the calculator functions defined
	// in "lib" and exports functions we'll use in this example.
	mod, err := r.InstantiateModuleFromCode(ctx, temperatureWasm)
	if err != nil {
		log.Panicln(err)
	}

	// Get references to WebAssembly functions we'll use in this example.
	fahrenheitToCelsius := mod.ExportedFunction("fahrenheitToCelsius")

	// Let's use the argument to this main function in Wasm.
	degreesStr := os.Args[1]
	degrees, err := strconv.Atoi(degreesStr)
	if err != nil {
		log.Panicln(err)
	}

	// The host object wasm refers to as externref.
	calculator := &calculator{}
	calculatorRef := uint64(uintptr(unsafe.Pointer(calculator)))

	results, err := fahrenheitToCelsius.Call(ctx, calculatorRef, uint64(degrees))
	if err != nil {
		log.Panicln(err)
	}

	fmt.Printf("%vF is %vC\n", degrees, results[0])
}

type calculator struct {
	stack []uint32
}

func _calculator_Push(calcPtr uintptr, val uint32) {
	calc := (*calculator)(unsafe.Pointer(calcPtr))
	calc.stack = append(calc.stack, val)
}

func _calculator_Pop(calcPtr uintptr) uint32 {
	calc := (*calculator)(unsafe.Pointer(calcPtr))
	num := len(calc.stack)
	if num == 0 {
		return 0
	}
	val := calc.stack[num-1]
	calc.stack = calc.stack[:num-1]
	return val
}

func _calculator_Add(calcPtr uintptr) {
	calc := (*calculator)(unsafe.Pointer(calcPtr))
	num := len(calc.stack)
	if num < 2 {
		return
	}
	a := calc.stack[num-2]
	b := calc.stack[num-1]
	res := a + b
	calc.stack = append(calc.stack[:num-2], res)
}

func _calculator_Subtract(calcPtr uintptr) {
	calc := (*calculator)(unsafe.Pointer(calcPtr))
	num := len(calc.stack)
	if num < 2 {
		return
	}
	a := calc.stack[num-2]
	b := calc.stack[num-1]
	res := a - b
	calc.stack = append(calc.stack[:num-2], res)
}

func _calculator_Multiply(calcPtr uintptr) {
	calc := (*calculator)(unsafe.Pointer(calcPtr))
	num := len(calc.stack)
	if num < 2 {
		return
	}
	a := calc.stack[num-2]
	b := calc.stack[num-1]
	res := a * b
	calc.stack = append(calc.stack[:num-2], res)
}

func _calculator_Divide(calcPtr uintptr) {
	calc := (*calculator)(unsafe.Pointer(calcPtr))
	num := len(calc.stack)
	if num < 2 {
		return
	}
	a := calc.stack[num-2]
	b := calc.stack[num-1]
	res := a / b
	calc.stack = append(calc.stack[:num-2], res)
}
