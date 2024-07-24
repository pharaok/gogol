package main

import (
	"math"
	"syscall/js"
)

func main() {
	universe := NewUniverse(8)

	document := js.Global().Get("document")
	canvasEl := document.Call("getElementById", "canvas")
	canvasEl.Set("tabIndex", 0)
	c := NewCanvas(canvasEl, universe)
	isPanning := false
	panX, panY := 0.0, 0.0

	canvasEl.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		e := args[0]
		gx, gy := c.ToGrid(e.Get("offsetX").Int(), e.Get("offsetY").Int())
		x, y := int(math.Floor(gx)), int(math.Floor(gy))
		gValue := universe.Get(x, y)
		universe.Set(x, y, gValue^1)
		return nil
	}))
	canvasEl.Call("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) any {
		e := args[0]
		switch e.Get("button").Int() {
		case 1:
			isPanning = true
			panX, panY = c.ToGrid(e.Get("offsetX").Int(), e.Get("offsetY").Int())
			e.Call("preventDefault")
		}
		return nil
	}))
	canvasEl.Call("addEventListener", "mousemove", js.FuncOf(func(this js.Value, args []js.Value) any {
		if isPanning {
			e := args[0]
			curX, curY := c.ToGrid(e.Get("offsetX").Int(), e.Get("offsetY").Int())
			c.originX += curX - panX
			c.originY += curY - panY
		}
		return nil
	}))
	canvasEl.Call("addEventListener", "mouseup", js.FuncOf(func(this js.Value, args []js.Value) any {
		e := args[0]
		switch e.Get("button").Int() {
		case 1:
			isPanning = false
		}
		return nil
	}))
	canvasEl.Call("addEventListener", "wheel", js.FuncOf(func(this js.Value, args []js.Value) any {
		e := args[0]

		delta := e.Get("deltaY").Float()
		factor := math.Pow(1.1, -delta/100) // magic numbers

		c.ZoomAt(factor, e.Get("offsetX").Int(), e.Get("offsetY").Int())
		return nil
	}))
	document.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
		e := args[0]
		// https://github.com/tinygo-org/tinygo/issues/1140
		switch e.Get("key").String() {
		case " ":
			universe.Step(3)
		}
		return nil
	}))

	channel := make(chan struct{})
	<-channel
}
