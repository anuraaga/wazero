package main

import "strings"

func main() {
	RegisterFunction(keepKeys, "keep_keys", "GetSetter", "[]string")
}

func notGetSetter() {
	panic("Not GetSetter")
}

func notStringSlice() {
	panic("Not []string")
}

// TinyGo does not support reflective invocation of functions so we do
// marshaling ourselves currently.
// https://github.com/tinygo-org/tinygo/blob/release/src/reflect/value.go#L865
func keepKeys(args []interface{}) ExprFunc {
	var target GetSetter
	var keys []string
	var ok bool

	if target, ok = args[0].(GetSetter); !ok {
		notGetSetter()
	}
	if keys, ok = args[1].([]string); !ok {
		notStringSlice()
	}

	for _, k := range keys {
		log(k)
	}

	keySet := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		keySet[key] = struct{}{}
	}

	return func(ctx TransformContext) interface{} {
		val := target.Get(ctx)
		if val == nil {
			return nil
		}

		if attrs, ok := val.(Map); ok {
			attrs.RemoveIf(func(key string, value Value) bool {
				_, ok := keySet[key]
				return !ok
			})
		}
		return nil
	}
}

// An extension function meant to allow arbitrary wasm to customize structs.
//export test
func test(mapPtr uint64) {
	m := &Map{ptr: mapPtr}

	m.RemoveIf(func(key string, value Value) bool {
		return strings.HasPrefix(key, "shark")
	})
}
