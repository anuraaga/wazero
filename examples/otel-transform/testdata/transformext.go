// This file is a standard library, it is maintained by the plugin architecture, not the user.

package main

import (
	"reflect"
	"unsafe"
)

type Value struct {
	ptr uint64
}

type Map struct {
	ptr uint64
}

//go:wasm-module oteltransform
//export Map_Len
func _Map_Len(mapPtr uint64) uint32
func (m Map) Len() uint32 {
	return _Map_Len(m.ptr)
}

//go:wasm-module oteltransform
//export Map_Range
func _Map_Range(mapPtr uint64) (itrPtr uint64)

//go:wasm-module oteltransform
//export Map_Range_Done
func _Map_Range_Done(itrPtr uint64) uint32

//go:wasm-module oteltransform
//export Map_Range_Key
func _Map_Range_Key(itrPtr uint64, bufOff uint32) (bufSize uint32)

//go:wasm-module oteltransform
//export Map_Range_Value
func _Map_Range_Value(itrPtr uint64) (valPtr uint64)

//go:wasm-module oteltransform
//export Map_Range_Advance
func _Map_Range_Advance(itrPtr uint64)

func (m Map) Range(f func(string, Value) bool) {
	itrPtr := _Map_Range(m.ptr)

	for _Map_Range_Done(itrPtr) == 0 {
		key := stringFromHost(itrPtr, func(ptr uint64, bufOff uint32) (size uint32) {
			//  cannot use an exported function as value:
			return _Map_Range_Key(ptr, bufOff)
		})
		val := Value{ptr: _Map_Range_Value(itrPtr)}
		f(key, val)
		_Map_Range_Advance(itrPtr)
	}
}

//go:wasm-module oteltransform
//export Map_RemoveIf
func _Map_RemoveIf(ptr uint64) (itrPtr uint64)

//go:wasm-module oteltransform
//export Map_RemoveIf_AddKey
func _Map_RemoveIf_AddKey(itrPtr uint64, keyOff uint32, keyLen uint32)

//go:wasm-module oteltransform
//export Map_RemoveIf_Finish
func _Map_RemoveIf_Finish(itrPtr uint64)

func (m Map) RemoveIf(f func(string, Value) bool) {
	remover := _Map_RemoveIf(m.ptr)
	m.Range(func(key string, value Value) bool {
		if f(key, value) {
			keyOff, keyLen := stringToPtr(key)
			_Map_RemoveIf_AddKey(remover, keyOff, keyLen)
		}
		return false
	})
	_Map_RemoveIf_Finish(remover)
}

type Resource struct {
	ptr uint64

	attributes Map
}

func (r Resource) Attributes() Map {
	return r.attributes
}

type InstrumentationScope struct {
	ptr uint64
}

//go:wasm-module oteltransform
//export InstrumentationScope_GetName
func _InstrumentationScope_GetName(scopePtr uint64, bufOff uint32) (bufSize uint32)
func (s InstrumentationScope) GetName() string {
	return stringFromHost(s.ptr, func(ptr uint64, bufOff uint32) (size uint32) {
		return _InstrumentationScope_GetName(ptr, bufOff)
	})
}

//go:wasm-module oteltransform
//export InstrumentationScope_GetVersion
func _InstrumentationScope_GetVersion(scopePtr uint64, bufOff uint32) (bufSize uint32)
func (s InstrumentationScope) GetVersion() string {
	return stringFromHost(s.ptr, func(ptr uint64, bufOff uint32) (size uint32) {
		return _InstrumentationScope_GetVersion(ptr, bufOff)
	})
}

type TransformContext struct {
	ptr uint64
}

func (ctx TransformContext) GetInstrumentationScope() InstrumentationScope {
	return InstrumentationScope{}
}

func (ctx TransformContext) GetResource() Resource {
	return Resource{}
}

type ExprFunc func(ctx TransformContext) interface{}

type GetSetter struct {
	ptr uint64
}

const RET_TYPE_MAP = 1

//go:wasm-module oteltransform
//export getSetter_Get
func getSetter_Get(ptr uint64, ctxPtr uint64, retTypeOff uint32) uint64
func (g *GetSetter) Get(ctx TransformContext) interface{} {
	retType := uint32(0)
	retTypeOff := uint32(uintptr(unsafe.Pointer(&retType)))
	valPtr := getSetter_Get(g.ptr, ctx.ptr, retTypeOff)
	switch retType {
	case RET_TYPE_MAP:
		return Map{ptr: valPtr}
	}
	panic("Unrecognized ret type")
}

//go:wasm-module oteltransform
//export getSetter_Set
func getSetter_Set() (ptr uint64, valPtr uint64)
func (g *GetSetter) Set(ctx TransformContext, val interface{}) {
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

func (f *functionDefinition) SetName(name string) *functionDefinition {
	nameOff, nameSize := stringToPtr(name)
	_functionDefinition_setName(f.ptr, nameOff, nameSize)
	return f
}

//go:wasm-module oteltransform
//export functionDefinition_setIdx
func _functionDefinition_setIdx(ptr uint64, idx uint32)

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

var factories []func(args []interface{}) ExprFunc
var initialized []ExprFunc

// reflect.TypeOf(f).Name() doesn't work in TinyGo yet it seems.
func RegisterFunction(f func(args []interface{}) ExprFunc, name string, paramTypes ...string) {
	idx := len(factories)
	def := newFunctionDefinition().
		SetName(name).
		SetParams(paramTypes...)
	_functionDefinition_setIdx(def.ptr, uint32(idx))
	_functionDefinition_register(def.ptr)
	factories = append(factories, f)
}

//go:wasm-module oteltransform
//export resolvedParam_done
func _resolvedParam_Done(itrPtr uint64) uint32

//go:wasm-module oteltransform
//export resolvedParam_advance
func _resolvedParam_Advance(itrPtr uint64)

//go:wasm-module oteltransform
//export resolvedParam_getType
func _resolvedParam_getType(itrPtr uint64, typeOff uint32) (size uint32)

//go:wasm-module oteltransform
//export resolvedParam_getPtr
func _resolvedParam_getPtr(itrPtr uint64) uint64

//export invoke_factory
func _invoke_factory(idx uint32, paramsItr uint64) uint32 {
	f := factories[idx]

	args := make([]interface{}, 0)
	for _resolvedParam_Done(paramsItr) != 1 {
		paramType := stringFromHost(paramsItr, func(ptr uint64, bufOff uint32) (size uint32) {
			return _resolvedParam_getType(ptr, bufOff)
		})
		paramPtr := _resolvedParam_getPtr(paramsItr)

		switch paramType {
		case "GetSetter":
			args = append(args, GetSetter{ptr: paramPtr})
		case "[]string":
			args = append(args, stringSliceFromHost(paramPtr))
		default:
			args = append(args, nil)
		}

		_resolvedParam_Advance(paramsItr)
	}
	initializedIdx := uint32(len(initialized))
	initialized = append(initialized, f(args))
	return initializedIdx
}

//export invoke_function
func invoke_function(idx uint32, ctxPtr uint64) {
	f := initialized[idx]
	ctx := TransformContext{ptr: ctxPtr}
	f(ctx)
}

// stringToPtr returns a offset and size pair for the given string
// in a way that is compatible with WebAssembly numeric types.
func stringToPtr(s string) (uint32, uint32) {
	buf := []byte(s)
	ptr := &buf[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	return uint32(unsafePtr), uint32(len(buf))
}

// Webassembly can only read from memory allocated within webassembly, so
// the host will have to copy data from the transformation context into
// a buffer we provide for reading strings.
func stringFromHost(ptr uint64, f func(ptr uint64, bufOff uint32) (size uint32)) string {
	buf := make([]byte, 0, 255)
	bufHdr := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	size := f(ptr, uint32(bufHdr.Data))
	return string(buf[:size])
}

//go:wasm-module oteltransform
//export StringSlice_Len
func _StringSlice_len(ptr uint64) uint32

//go:wasm-module oteltransform
//export StringSlice_Get
func _StringSlice_get(ptr uint64, i uint32, bufOff uint32) (size uint32)

//go:wasm-module oteltransform
//export log
func _log(off uint32, size uint32)

func log(s string) {
	off, size := stringToPtr(s)
	_log(off, size)
}

func stringSliceFromHost(ptr uint64) []string {
	buf := make([]byte, 0, 255)
	bufHdr := (*reflect.SliceHeader)(unsafe.Pointer(&buf))

	num := int(_StringSlice_len(ptr))

	ret := make([]string, num)
	for i := 0; i < num; i++ {
		size := _StringSlice_get(ptr, uint32(i), uint32(bufHdr.Data))
		ret[i] = string(buf[:size])
	}

	return ret
}
