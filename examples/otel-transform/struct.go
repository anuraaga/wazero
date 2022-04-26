package main

import (
	"context"
	_ "embed"
	"fmt"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"log"
	"unsafe"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/wasi"
)

// structWasm was compiled using `tinygo build -o struct.wasm -scheduler=none --no-debug -target=wasi struct.go`
//go:embed testdata/lib.wasm
var structWasm []byte

// See README.md for a full description.
func main() {
	// Choose the context to use for function calls.
	ctx := context.Background()

	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntimeWithConfig(wazero.NewRuntimeConfigInterpreter())

	// Instantiate a module named "env" that exports operations on a bear.
	env, err := r.NewModuleBuilder("oteltransform").
		ExportFunction("log", _log).
		ExportFunction("Map_Len", map_Len).
		ExportFunction("Map_Range", map_Range).
		ExportFunction("Map_Range_Done", map_Range_Done).
		ExportFunction("Map_Range_Key", map_Range_Key).
		ExportFunction("Map_Range_Value", map_Range_Value).
		ExportFunction("Map_Range_Advance", map_Range_Advance).
		ExportFunction("Map_RemoveIf", map_RemoveIf).
		ExportFunction("Map_RemoveIf_AddKey", map_RemoveIf_AddKey).
		ExportFunction("Map_RemoveIf_Finish", map_RemoveIf_Finish).
		ExportFunction("newFunctionDefinition", newFunctionDefinition).
		ExportFunction("functionDefinition_setName", functionDefinition_setName).
		ExportFunction("functionDefinition_clearParams", functionDefinition_clearParams).
		ExportFunction("functionDefinition_addParam", functionDefinition_addParam).
		ExportFunction("functionDefinition_setIdx", functionDefinition_setIdx).
		ExportFunction("functionDefinition_register", functionDefinition_register).
		ExportFunction("resolvedParam_done", resolvedParam_done).
		ExportFunction("resolvedParam_advance", resolvedParam_advance).
		ExportFunction("resolvedParam_getType", resolvedParam_getType).
		ExportFunction("resolvedParam_getPtr", resolvedParam_getPtr).
		ExportFunction("getSetter_Get", getSetter_Get).
		ExportFunction("StringSlice_Len", stringSlice_len).
		ExportFunction("StringSlice_Get", stringSlice_get).
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

	invoke_factory := mod.ExportedFunction("invoke_factory")
	invoke_function := mod.ExportedFunction("invoke_function")

	gs := &testGetSetter{}
	keys := &[]string{"bear.name", "bear.greeting"}
	params := []resolvedParam{
		{typ: "GetSetter", ptr: toPtr(unsafe.Pointer(gs))},
		{typ: "[]string", ptr: toPtr(unsafe.Pointer(keys))},
	}
	paramsItr := &resolvedParamIterator{
		params: params,
		pos:    0,
	}

	res, err := invoke_factory.Call(ctx, uint64(0), toPtr(unsafe.Pointer(paramsItr)))
	if err != nil {
		log.Fatal(err)
	}

	m := pcommon.NewMap()
	m.InsertString("bear.name", "Yogi")
	m.InsertString("bear.greeting", "Yabba Dabba Doo!")
	m.InsertString("shark.name", "Jaws")
	m.InsertString("shark.greeting", "Chomp!")
	transformCtx := &testTransformContext{
		m: m,
	}

	_, err = invoke_function.Call(ctx, res[0], toPtr(unsafe.Pointer(transformCtx)))
	if err != nil {
		log.Fatal(err)
	}

	// Name should have been updated in wasm
	m.Range(func(key string, val pcommon.Value) bool {
		fmt.Printf("%v:%v\n", key, val.AsString())
		return true
	})
}

func toPtr(val unsafe.Pointer) uint64 {
	return uint64(uintptr(val))
}

type testTransformContext struct {
	m pcommon.Map
}

func (t *testTransformContext) GetItem() interface{} {
	//TODO implement me
	panic("implement me")
}

func (t *testTransformContext) GetInstrumentationScope() pcommon.InstrumentationScope {
	//TODO implement me
	panic("implement me")
}

func (t *testTransformContext) GetResource() pcommon.Resource {
	//TODO implement me
	panic("implement me")
}

type testGetSetter struct {
}

func (g *testGetSetter) Get(ctx TransformContext) interface{} {
	return &ctx.(*testTransformContext).m
}

func (g *testGetSetter) Set(ctx TransformContext, val interface{}) {
}
