package main

import (
	"syscall/js"
)

func main() {
	universe := NewUniverse(10)

	document := js.Global().Get("document")
	canvasEl := document.Call("getElementById", "canvas")
	canvasEl.Set("tabIndex", 0)
	c := NewCanvas(canvasEl, universe)

	canvasEl.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		e := args[0]
		x, y := e.Get("offsetX").Int(), e.Get("offsetY").Int()
		gx, gy := int(float64(x)/c.cellSize), int(float64(y)/c.cellSize)
		gValue := universe.Get(gx, gy)
		universe.Set(gx, gy, gValue^1)
		return nil
	}))
	document.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
		universe.Step(0)
		return nil
	}))

	channel := make(chan struct{})
	<-channel
}
