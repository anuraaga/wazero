package main

import (
	"fmt"
	"github.com/tetratelabs/wazero/api"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"log"
	"reflect"
	"unsafe"
)

// Stores iterators while used in wasm to prevent GC. All iterators require some terminal operation
// that removes from this store.
var store = make(map[interface{}]struct{})

var wasmFunctions = make(map[string]functionDefinition)

func map_Len(ptr uint64) uint32 {
	m := (*pcommon.Map)(toUnsafePointer(ptr))
	return uint32(m.Len())
}

type keyValue struct {
	key   string
	value pcommon.Value
}

type rangeIterator struct {
	items []keyValue
	pos   int
}

func map_Range(ptr uint64) uint64 {
	m := (*pcommon.Map)(toUnsafePointer(ptr))

	itr := &rangeIterator{items: make([]keyValue, 0, m.Len()), pos: 0}
	m.Range(func(key string, val pcommon.Value) bool {
		itr.items = append(itr.items, keyValue{key: key, value: val})
		return true
	})

	store[itr] = struct{}{}
	return toPtr(unsafe.Pointer(itr))
}

func map_Range_Done(itrPtr uint64) uint32 {
	itr := (*rangeIterator)(toUnsafePointer(itrPtr))

	res := itr.pos == len(itr.items)
	// HACK - the iterator cannot be used after done returns true.
	// We know this is how our wasm runtime is written but it may
	// be too unsafe. Or maybe it's OK compared to everything else
	// we do.
	if res {
		delete(store, itr)
		return 1
	}
	return 0
}

func map_Range_Key(m api.Module, itrPtr uint64, bufOff uint32) uint32 {
	itr := (*rangeIterator)(toUnsafePointer(itrPtr))
	key := itr.items[itr.pos].key
	if !m.Memory().Write(bufOff, []byte(key)) {
		log.Fatalf("Memory.Write(%d, %d) out of range of memory size %d",
			bufOff, len(key), m.Memory().Size())
	}
	return uint32(len(key))
}

func map_Range_Value(itrPtr uint64) uint64 {
	itr := (*rangeIterator)(toUnsafePointer(itrPtr))
	value := itr.items[itr.pos].value
	return toPtr(unsafe.Pointer(&value))
}

func map_Range_Advance(itrPtr uint64) {
	itr := (*rangeIterator)(toUnsafePointer(itrPtr))
	itr.pos++
}

type removeIterator struct {
	m        *pcommon.Map
	toRemove map[string]struct{}
}

func map_RemoveIf(ptr uint64) uint64 {
	m := (*pcommon.Map)(toUnsafePointer(ptr))

	itr := &removeIterator{m: m, toRemove: make(map[string]struct{})}
	store[itr] = struct{}{}
	return uint64(uintptr(unsafe.Pointer(itr)))
}

func map_RemoveIf_AddKey(m api.Module, itrPtr uint64, keyOff uint32, keyLen uint32) {
	buf, ok := m.Memory().Read(keyOff, keyLen)
	if !ok {
		log.Fatalf("Memory.Read(%d, %d) out of range", keyOff, keyLen)
	}

	itr := (*removeIterator)(toUnsafePointer(itrPtr))
	itr.toRemove[string(buf)] = struct{}{}
}

func map_RemoveIf_Finish(itrPtr uint64) {
	itr := (*removeIterator)(toUnsafePointer(itrPtr))
	itr.m.RemoveIf(func(key string, value pcommon.Value) bool {
		_, ok := itr.toRemove[key]
		return ok
	})
	delete(store, itr)
}

type functionDefinition struct {
	name   string
	params []string
	idx    uint32
}

func newFunctionDefinition() uint64 {
	fd := &functionDefinition{}
	store[fd] = struct{}{}
	return toPtr(unsafe.Pointer(fd))
}

func functionDefinition_setName(m api.Module, ptr uint64, nameOff uint32, nameSize uint32) {
	fd := (*functionDefinition)(toUnsafePointer(ptr))
	buf, ok := m.Memory().Read(nameOff, nameSize)
	if !ok {
		log.Fatalf("Memory.Read(%d, %d) out of range", nameOff, nameSize)
	}
	fd.name = string(buf)
}

func functionDefinition_clearParams(ptr uint64) {
	fd := (*functionDefinition)(toUnsafePointer(ptr))
	fd.params = nil
}

func functionDefinition_addParam(m api.Module, ptr uint64, nameOff uint32, nameSize uint32) {
	fd := (*functionDefinition)(toUnsafePointer(ptr))
	buf, ok := m.Memory().Read(nameOff, nameSize)
	if !ok {
		log.Fatalf("Memory.Read(%d, %d) out of range", nameOff, nameSize)
	}
	fd.params = append(fd.params, string(buf))
}

func functionDefinition_setIdx(ptr uint64, idx uint32) {
	fd := (*functionDefinition)(toUnsafePointer(ptr))
	fd.idx = idx
}

func functionDefinition_register(ptr uint64) {
	fd := (*functionDefinition)(toUnsafePointer(ptr))
	wasmFunctions[fd.name] = *fd
	delete(store, fd)
}

type resolvedParam struct {
	typ string
	ptr uint64
}

type resolvedParamIterator struct {
	params []resolvedParam
	pos    int
}

func resolvedParam_getType(m api.Module, itrPtr uint64, typeOff uint32) uint32 {
	itr := (*resolvedParamIterator)(toUnsafePointer(itrPtr))
	t := itr.params[itr.pos].typ
	if !m.Memory().Write(typeOff, []byte(t)) {
		log.Fatalf("Memory.Write(%d, %d) out of range of memory size %d",
			typeOff, len(t), m.Memory().Size())
	}
	return uint32(len(t))
}

func resolvedParam_getPtr(itrPtr uint64) uint64 {
	itr := (*resolvedParamIterator)(toUnsafePointer(itrPtr))
	return itr.params[itr.pos].ptr
}

func resolvedParam_done(itrPtr uint64) uint32 {
	itr := (*resolvedParamIterator)(toUnsafePointer(itrPtr))
	res := itr.pos == len(itr.params)
	if res {
		return 1
	}
	return 0
}

func resolvedParam_advance(itrPtr uint64) {
	itr := (*resolvedParamIterator)(toUnsafePointer(itrPtr))
	itr.pos++
}

func stringSlice_len(ptr uint64) uint32 {
	ss := (*[]string)(toUnsafePointer(ptr))
	return uint32(len(*ss))
}

func stringSlice_get(m api.Module, ptr uint64, i uint32, bufOff uint32) uint32 {
	ss := (*[]string)(toUnsafePointer(ptr))
	s := (*ss)[i]
	if !m.Memory().Write(bufOff, []byte(s)) {
		log.Fatalf("Memory.Write(%d, %d) out of range of memory size %d",
			bufOff, len(s), m.Memory().Size())
	}
	return uint32(len(s))
}

const RET_TYPE_MAP = 1

// https://stackoverflow.com/questions/7132848/how-to-get-the-reflect-type-of-an-interface
var getSetterType = reflect.TypeOf((*GetSetter)(nil)).Elem()
var transformContextType = reflect.TypeOf((*TransformContext)(nil)).Elem()

func getSetter_Get(m api.Module, ptr uint64, ctxPtr uint64, retTypeOff uint32) uint64 {
	//gs := reflect.NewAt(getSetterType, toUnsafePointer(ptr)).Interface().(*GetSetter)
	//ctx := reflect.NewAt(transformContextType, toUnsafePointer(ctxPtr)).Interface().(*TransformContext)

	//gs := (*GetSetter)(toUnsafePointer(ptr))
	//ctx := (*TransformContext)(toUnsafePointer(ctxPtr))

	gs := (*testGetSetter)(toUnsafePointer(ptr))
	ctx := (*testTransformContext)(toUnsafePointer(ctxPtr))

	val := (*gs).Get(ctx)
	switch c := val.(type) {
	case *pcommon.Map:
		m.Memory().WriteUint32Le(retTypeOff, RET_TYPE_MAP)
		return toPtr(unsafe.Pointer(c))
	default:
		panic("Bad type")
	}
}

func _log(m api.Module, off uint32, size uint32) {
	buf, ok := m.Memory().Read(off, size)
	if !ok {
		log.Fatalf("Memory.Read(%d, %d) out of range", off, size)
	}
	fmt.Println(string(buf))
}

func toUnsafePointer(ptr uint64) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr))
}
