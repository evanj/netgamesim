package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"syscall/js"

	"github.com/llgcode/draw2d"
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

	// spriteImgs := sprites.New()
	// draw.Draw(screen.frame, image.Rect(200, 200, 200+12, 200+12),
	// 	&spriteImgs.Tank, image.Point{0, 0}, draw.Over)

	a := &animation{0, screen, js.Func{}}
	a.drawFrameJSFunc = js.FuncOf(a.drawFrame)
	// a.drawFrame(0, centerX, centerY)
	// a.drawFrame(0, centerX+2*size+0.5, centerY+0.5)
	// a.d1px(a.step, centerX+2*size, centerY+2*size)

	// Render the frame
	// screen.renderFrame()

	js.Global().Set("goStepCallback", js.FuncOf(a.stepCallback))

	//

	// // TODO: call drawFrameJSFunc.Release()
	js.Global().Call("requestAnimationFrame", a.drawFrameJSFunc)

	done := make(chan struct{})
	<-done
}

func (a *animation) drawFrame(this js.Value, args []js.Value) interface{} {
	// called with DOMHighResTimeStamp indicating seconds since document start; ignored
	// log.Printf("requestAnimationFrame args[0]=%v", args[0])
	msSinceStart := args[0].Float()
	secondsSinceStart := msSinceStart / 1000.

	rotations := secondsSinceStart * rotationsPerSecond
	fractional := rotations - math.Floor(rotations)
	a.drawStroke(fractional, centerX, centerY)
	a.drawStroke(fractional, centerX+2*size+0.5, centerY+0.5)
	a.d1px(fractional, centerX+2*size, centerY+2*size)

	a.screen.renderFrame()
	js.Global().Call("requestAnimationFrame", a.drawFrameJSFunc)
	log.Printf("wtf %f frac %f", msSinceStart, fractional)

	return nil
}

func (a *animation) stepCallback(this js.Value, args []js.Value) interface{} {
	// called with DOMHighResTimeStamp indicating seconds since document start; ignored
	// a.step += 1
	// a.drawFrame(a.step, centerX, centerY)
	// a.drawFrame(a.step, centerX+2*size+0.5, centerY+0.5)
	// a.d1px(a.step, centerX+2*size, centerY+2*size)

	// a.screen.renderFrame()
	return nil
}

type animation struct {
	// start time.Time
	step            int
	screen          *canvasScreen
	drawFrameJSFunc js.Func
}

const centerX = 100.0
const centerY = 100.0
const size = 50.0
const rotationsPerSecond = 0.5

func (a *animation) drawStroke(fraction float64, centerX float64, centerY float64) {
	a.screen.gc.SetStrokeColor(color.Black)
	a.screen.gc.SetLineWidth(1.0)
	a.screen.gc.BeginPath()
	a.screen.gc.MoveTo(centerX, centerY)

	// compute the end point
	rads := 2 * math.Pi * fraction
	yOffset := math.Sin(rads) * size
	xOffset := math.Cos(rads) * size
	// log.Printf("step=%d x=%f y=%f", xOffset, yOffset)
	a.screen.gc.LineTo(centerX+xOffset, centerY+yOffset)
	a.screen.gc.Stroke()

	// log.Printf("fraction=%d rads=%f; draw line %f,%f -> %f,%f",
	// 	step, rads, float64(centerX), float64(centerY), centerX+xOffset, centerY+yOffset)
}

func (a *animation) d1px(fraction float64, centerX int, centerY int) {
	// compute the end point
	rads := 2 * math.Pi * fraction
	yOffset := math.Sin(rads) * size
	xOffset := math.Cos(rads) * size

	x := centerX + int(xOffset+0.5)
	y := centerY + int(yOffset+0.5)
	draw1PxLine(a.screen.gc, image.Point{int(centerX), int(centerY)}, image.Point{x, y}, color.Black)
}

func iabs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// draw1Px draws a single pixel think linkthe rectangle described by image.Rectangle with the given color
func draw1PxLine(gc draw2d.GraphicContext, p1 image.Point, p2 image.Point, c color.Color) {
	// determine the "primary" direction as the largest magnitude
	xDiff := p2.X - p1.X
	yDiff := p2.Y - p1.Y

	xStartOffset := 0.5
	xEndOffset := 0.5
	yStartOffset := 0.5
	yEndOffset := 0.5
	if iabs(xDiff) > iabs(yDiff) {
		xStartOffset = 0
		xEndOffset = 1
		if xDiff < 0 {
			xStartOffset = 1
			xEndOffset = 0
		}
	} else if iabs(xDiff) < iabs(yDiff) {
		yStartOffset = 0
		yEndOffset = 1
		if yDiff < 0 {
			yStartOffset = 1
			yEndOffset = 0
		}
	}

	gc.SetStrokeColor(c)
	gc.BeginPath()
	gc.MoveTo(float64(p1.X)+xStartOffset, float64(p1.Y)+yStartOffset)
	gc.LineTo(float64(p2.X)+xEndOffset, float64(p2.Y)+yEndOffset)
	gc.Stroke()
}

func newScreen(canvasElement js.Value) *canvasScreen {
	width := int(canvasElement.Get("width").Float())
	height := int(canvasElement.Get("height").Float())

	// Create a Go image to draw into
	frame := image.NewRGBA(image.Rect(0, 0, width, height))
	gc := draw2dimg.NewGraphicContext(frame)

	// set up the Canvas so we can draw to it
	ctx := canvasElement.Call("getContext", "2d")
	imageData := ctx.Call("createImageData", js.ValueOf(width), js.ValueOf(height))
	imageDataData := imageData.Get("data")
	jsUInt8Array := js.Global().Get("Uint8Array").New(len(frame.Pix))

	return &canvasScreen{width, height, ctx, imageData, imageDataData, jsUInt8Array, frame, gc}
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

	frame *image.RGBA
	gc    *draw2dimg.GraphicContext
}

func (c *canvasScreen) renderFrame() {
	// ImageData.data is a UInt8ClampedArray; we can't copy directly to it
	// copy to a UInt8Array, then set that
	n := js.CopyBytesToJS(c.jsUInt8Array, c.frame.Pix)
	if n != len(c.frame.Pix) {
		panic(fmt.Sprintf("should be impossible: copy failed %d %d", n, len(c.frame.Pix)))
	}
	c.imageDataData.Call("set", c.jsUInt8Array)
	c.jsCtx.Call("putImageData", c.imageData, 0, 0)

	// erase the frame memory with transparent pixels
	for i := range c.frame.Pix {
		c.frame.Pix[i] = 0
	}
}
