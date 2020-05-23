package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"syscall/js"

	"github.com/llgcode/draw2d/draw2dimg"
)

const canvasID = "canvas"

func main() {
	log.Printf("demo loading in canvas id=%s ...", canvasID)

	// locate the canvas
	document := js.Global().Get("document")
	canvasElement := document.Call("getElementById", canvasID)
	screen := newScreen(canvasElement)

	log.Printf("canvas dimensions width:%d x height:%d", screen.width, screen.height)

	// Set some properties
	screen.gc.SetFillColor(color.RGBA{0x44, 0xff, 0x44, 0xff})
	screen.gc.SetStrokeColor(color.RGBA{0x44, 0x44, 0x44, 0xff})
	screen.gc.SetLineWidth(5)

	// Draw a closed shape
	screen.gc.MoveTo(10, 10) // should always be called first for a new path
	screen.gc.LineTo(100, 50)
	screen.gc.QuadCurveTo(100, 10, 10, 10)
	screen.gc.Close()
	screen.gc.FillStroke()

	// Render the frame
	screen.renderFrame()
}

type animation struct {
	screen *canvasScreen
	x      int
	y      int
}

func newScreen(canvasElement js.Value) *canvasScreen {
	width := int(canvasElement.Get("width").Float())
	height := int(canvasElement.Get("height").Float())

	// Create a Go image to draw into
	drawImage := image.NewRGBA(image.Rect(0, 0, width, height))
	gc := draw2dimg.NewGraphicContext(drawImage)

	// set up the Canvas so we can draw to it
	ctx := canvasElement.Call("getContext", "2d")
	imageData := ctx.Call("createImageData", js.ValueOf(width), js.ValueOf(height))
	imageDataData := imageData.Get("data")
	jsUInt8Array := js.Global().Get("Uint8Array").New(len(drawImage.Pix))

	return &canvasScreen{width, height, ctx, imageData, imageDataData, jsUInt8Array, drawImage, gc}
	// c.image = image.NewRGBA(image.Rect(0, 0, width, height))
	// c.copybuff = js.Global().Get("Uint8Array").New(len(c.image.Pix)) // Static JS buffer for copying data out to JS. Defined once and re-used to save on un-needed allocations
}

type canvasScreen struct {
	width  int
	height int

	jsCtx         js.Value
	imageData     js.Value
	imageDataData js.Value
	jsUInt8Array  js.Value

	buffer *image.RGBA
	gc     *draw2dimg.GraphicContext
}

func (c *canvasScreen) renderFrame() {
	// ImageData.data is a UInt8ClampedArray; we can't copy directly to it
	// copy to a UInt8Array, then set that
	n := js.CopyBytesToJS(c.jsUInt8Array, c.buffer.Pix)
	if n != len(c.buffer.Pix) {
		panic(fmt.Sprintf("should be impossible: copy failed %d %d", n, len(c.buffer.Pix)))
	}
	c.imageDataData.Call("set", c.jsUInt8Array)
	c.jsCtx.Call("putImageData", c.imageData, 0, 0)
}
