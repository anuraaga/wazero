package main

import (
	"unsafe"
)

type Value interface {
}

type Map struct {
	ptr uint64
}

func (m Map) RemoveIf(f func(string, Value) bool) {

}

type Resource struct {
	ptr uint64

	attributes Map
}

func (r Resource) Attributes() Map {

}

type InstrumentationScope struct {
	ptr uint64
}

//go:wasm-module oteltransform
//export InstrumentationScope_GetName
func _InstrumentationScope_GetName() string

func (s InstrumentationScope) GetVersion() string {

}

type TransformContext struct {
	ptr uint64

	scope    InstrumentationScope
	resource Resource
}

func (ctx TransformContext) GetInstrumentationScope() InstrumentationScope {
	return ctx.scope
}

func (ctx TransformContext) GetResource() Resource {
	return ctx.resource
}

type ExprFunc func(ctx TransformContext) interface{}

type Getter interface {
	Get(ctx TransformContext) interface{}
}

type Setter interface {
	Set(ctx TransformContext, val interface{})
}

type GetSetter interface {
	Getter
	Setter
}

type getSetter struct {
	ptr uint64
}

func (g *getSetter) Get(ctx TransformContext) interface{} {
	//TODO implement me
	panic("implement me")
}

func (g *getSetter) Set(ctx TransformContext, val interface{}) {
	//TODO implement me
	panic("implement me")
}

type functionDefinition struct {
	ptr uint64
}

//go:wasm-module oteltransform
//export newFunctionDefinition
func _newFunctionDefinition() (ptr uint64)
func newFunctionDefinition() *functionDefinition {
	return &functionDefinition{ptr: _newFunctionDefinition()}
}

//go:wasm-module oteltransform
//export functionDefinition_setName

func _functionDefinition_setName(ptr uint64, off uint32, size uint32)

func (f *functionDefinition) SetName(s string) *functionDefinition {
	return f
}

//go:wasm-module oteltransform
//export functionDefinition_clearParams
func _functionDefinition_clearParams(ptr uint64)

//go:wasm-module oteltransform
//export functionDefinition_addParam
func _functionDefinition_addParam(ptr uint64, off uint32, size uint32)

func (f *functionDefinition) SetParams(params ...string) *functionDefinition {
	_functionDefinition_clearParams(f.ptr)
	for _, p := range params {
		pOff, pSize := stringToPtr(p)
		_functionDefinition_addParam(f.ptr, pOff, pSize)
	}

	return f
}

//go:wasm-module oteltransform
//export functionDefinition_register
func _functionDefinition_register(ptr uint64)

func RegisterFunction(name string, params ...string) {
	f := newFunctionDefinition().
		SetName(name).
		SetParams(params...)
	_functionDefinition_register(f.ptr)
}

// stringToPtr returns a offset and size pair for the given string
// in a way that is compatible with WebAssembly numeric types.
func stringToPtr(s string) (uint32, uint32) {
	buf := []byte(s)
	ptr := &buf[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	return uint32(unsafePtr), uint32(len(buf))
}
