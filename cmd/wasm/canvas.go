//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/pharaok/gogol/pkg/hashlife"
)

type Canvas struct {
	ctx              js.Value
	originX, originY float64
	cellSize         float64
	universe         *hashlife.Universe
}

const LeafLevel = hashlife.LeafLevel
const LeafHalfSize = hashlife.LeafHalfSize
const LeafSize = hashlife.LeafSize

func NewCanvas(canvasEl js.Value, universe *hashlife.Universe) *Canvas {
	window := js.Global().Get("window")
	width, height := window.Get("innerWidth").Int(), window.Get("innerHeight").Int()

	ctx := canvasEl.Call("getContext", "2d")
	canvasEl.Set("width", width)
	canvasEl.Set("height", height)

	canvas := &Canvas{ctx: ctx, cellSize: 20.0, universe: universe}

	var animate js.Func
	animate = js.FuncOf(func(this js.Value, args []js.Value) any {
		window.Call("requestAnimationFrame", animate)
		canvas.Paint()
		return nil
	})
	window.Call("requestAnimationFrame", animate)

	return canvas
}

func (c *Canvas) ToGrid(x, y int) (float64, float64) {
	return float64(x)/c.cellSize - c.originX, float64(y)/c.cellSize - c.originY
}

func (c *Canvas) Clear() {
	canvasEl := c.ctx.Get("canvas")
	width, height := canvasEl.Get("width").Int(), canvasEl.Get("height").Int()
	c.ctx.Call("clearRect", 0, 0, width, height)
}
func (c *Canvas) PaintNode(n *hashlife.Node, left, top float64) {
	if n == nil {
		return
	}

	if n.Level == LeafLevel {
		c.ctx.Set("fillStyle", "black")
		for y := 0; y < LeafSize; y++ {
			for x := 0; x < LeafSize; x++ {
				if n.Get(x-LeafHalfSize, y-LeafHalfSize) == 1 {
					sz := c.cellSize
					c.ctx.Call("fillRect", (left+float64(x))*sz, (top+float64(y))*sz, sz, sz)
				}
			}
		}
		return
	}

	halfSize := float64(int(1) << (n.Level - 1))

	c.PaintNode(n.Child(-1, -1), left, top)
	c.PaintNode(n.Child(0, -1), left+halfSize, top)
	c.PaintNode(n.Child(-1, 0), left, top+halfSize)
	c.PaintNode(n.Child(0, 0), left+halfSize, top+halfSize)
}
func (c *Canvas) Paint() {
	halfSize := float64(int(1) << (c.universe.Root.Level - 1))
	c.Clear()
	c.PaintNode(c.universe.Root, c.originX-halfSize, c.originY-halfSize)
}

func (c *Canvas) ZoomAt(factor float64, x, y int) {
	preX, preY := c.ToGrid(x, y)
	c.cellSize *= factor
	postX, postY := c.ToGrid(x, y)

	c.originX += postX - preX
	c.originY += postY - preY
}
