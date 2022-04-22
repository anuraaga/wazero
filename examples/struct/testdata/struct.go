package main

// main is required for TinyGo to compile to Wasm.
func main() {}

// An extension function meant to allow arbitrary wasm to customize structs.
//export hello
func _hello(bearPtr uint64) {
	// First step is always to switch to the convenience runtime API. Ideally this could happen automatically if
	// the parameter was of type Bear instead of uint64
	b := WrapBear(bearPtr)
	greeted := "Hello " + b.GetName()
	b.SetName(greeted)
}
