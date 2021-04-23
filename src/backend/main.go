package main

import (
	"fmt"
	"syscall/js"
)

func main() {
	println("Go WebAssembly Initialized")

	js.Global().Set("render", js.FuncOf(render))

	<-make(chan bool)
}

func render(this js.Value, args []js.Value) interface{} {
	println("render called 2")
	fmt.Printf("%v\n", this.Get("name"))
	fmt.Printf("%v\n", args)
	return nil
}
