package main

import (
	"reflect"
	"unsafe"
)

// This file is the runtime library a webassembly extension would be compiled against if implementing
// operations that work on host data. It's main job is to marshal data back and forth between host
// memory and wasm.

//go:wasm-module env
//export setName
func _setName(bearPtr uint64, nameOff uint32, nameLen uint32)

//go:wasm-module env
//export getName
func _getName(bearPtr uint64, bufOff uint32) (size uint32)

type Bear interface {
	GetName() string
	SetName(string)
}

func WrapBear(ptr uint64) Bear {
	return &goBear{ptr: ptr}
}

type goBear struct {
	ptr uint64
}

func (b *goBear) GetName() string {
	buf := make([]byte, 0, 255)

	bufHdr := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	nameLen := _getName(b.ptr, uint32(bufHdr.Data))

	return string(buf[:nameLen])
}

func (b *goBear) SetName(name string) {
	off, size := stringToPtr(name)
	_setName(b.ptr, off, size)
}

// stringToPtr returns a offset and size pair for the given string
// in a way that is compatible with WebAssembly numeric types.
func stringToPtr(s string) (uint32, uint32) {
	buf := []byte(s)
	ptr := &buf[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	return uint32(unsafePtr), uint32(len(buf))
}
