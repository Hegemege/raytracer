package main

import "syscall/js"

func main() {
	//c := make(chan struct{}, 0)

	println("Go WebAssembly Initialized")

	js.Global().Set("render", js.FuncOf(render))

	//<-c
}

func render(this js.Value, args []js.Value) interface{} {
	println("test")
	return nil
}
