// +build wasm

package main

import (
	"fmt"
	"image"
	"log"
	"syscall/js"

	"github.com/evanj/netgamesim/game"
	"github.com/evanj/netgamesim/sprites"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

const canvasID = "canvas"

const keyCodeSpace = 32
const keyCodeLeft = 37
const keyCodeUp = 38
const keyCodeRight = 39
const keyCodeDown = 40

const logFPSSeconds = 15

type client struct {
	screen *canvasScreen

	keyDownCallback js.Func
	keyUpCallback   js.Func
	requestFrame    js.Func
	simTimeStart    float64
	lastFPSLogTime  float64
	frames          int

	game *game.Game

	firePressed bool
	tankDir     game.Direction
}

func newClient(screen *canvasScreen, g *game.Game) *client {
	c := &client{
		screen,
		js.Func{}, js.Func{}, js.Func{},
		0.0, 0.0, 0,
		g,
		false,
		game.DirNone,
	}
	c.keyDownCallback = js.FuncOf(c.jsKeyDown)
	c.keyUpCallback = js.FuncOf(c.jsKeyUp)
	c.requestFrame = js.FuncOf(c.jsRequestFrame)
	return c
}

func (c *client) Stop() {
	c.keyDownCallback.Release()
	c.keyUpCallback.Release()
	c.requestFrame.Release()
}

func dirFromKeyCode(keyCode int) game.Direction {
	switch keyCode {
	case keyCodeLeft:
		return game.DirLeft
	case keyCodeUp:
		return game.DirUp
	case keyCodeRight:
		return game.DirRight
	case keyCodeDown:
		return game.DirDown
	default:
		return game.DirNone
	}
}

func (c *client) jsKeyDown(this js.Value, args []js.Value) interface{} {
	event := args[0]
	// TODO: check for repeat?
	// repeat := event.Get("repeat").Bool()
	// probably not worth it if this involves a "call" back to the browser?

	input := game.Input{
		TankDir: c.tankDir,
		Fire:    false,
	}

	keyCode := event.Get("keyCode").Int()
	switch keyCode {
	case keyCodeSpace:
		if c.firePressed {
			// ignore duplicate
			event.Call("preventDefault")
			return nil
		}
		c.firePressed = true
		input.Fire = true

	case keyCodeLeft, keyCodeUp, keyCodeRight, keyCodeDown:
		dir := dirFromKeyCode(keyCode)
		if dir == game.DirNone {
			panic("BUG: mismatch between case and dirFromKeyCode")
		}
		input.TankDir = dir
		c.tankDir = input.TankDir

	default:
		// unknown key: ignore
		return nil
	}

	c.game.ProcessInput(input)

	// prevent keys from doing what they normally would
	event.Call("preventDefault")
	return nil
}

func (c *client) jsKeyUp(this js.Value, args []js.Value) interface{} {
	event := args[0]
	keyCode := event.Get("keyCode").Int()

	if keyCode == keyCodeSpace {
		c.firePressed = false
		return nil
	}

	dir := dirFromKeyCode(keyCode)
	if dir == game.DirNone {
		return nil
	}
	if dir == c.tankDir {
		c.tankDir = game.DirNone
		c.game.ProcessInput(game.Input{TankDir: game.DirNone, Fire: false})
	}
	return nil
}

func drawGame(gc draw2d.GraphicContext, g *game.Game) {
	sprites.DrawTank(gc, g.TankCenter())
	sprites.DrawTarget(gc, g.TargetCenter())
	for _, b := range g.Bullets() {
		sprites.DrawBullet(gc, b)
	}
	for _, s := range g.Smoke() {
		sprites.DrawSmoke(gc, s)
	}
}

func (c *client) jsRequestFrame(this js.Value, args []js.Value) interface{} {
	msSinceDocStart := args[0].Float()
	if c.simTimeStart == 0.0 {
		c.simTimeStart = msSinceDocStart
		c.lastFPSLogTime = msSinceDocStart
	}

	c.game.AdvanceSimulation(msSinceDocStart - c.simTimeStart)

	// draw the state of the universe
	drawGame(c.screen.gc, c.game)
	c.screen.renderFrame()

	// request the next frame
	js.Global().Call("requestAnimationFrame", c.requestFrame)

	c.frames++
	if msSinceDocStart-c.lastFPSLogTime >= logFPSSeconds*1000 {
		seconds := (msSinceDocStart - c.lastFPSLogTime) / 1000.0
		fps := float64(c.frames) / seconds
		log.Printf("t=%f frames=%d seconds=%f fps=%f", msSinceDocStart, c.frames, seconds, fps)
		c.frames = 0
		c.lastFPSLogTime = msSinceDocStart
	}
	return nil
}

func main() {
	log.Printf("demo loading in canvas id=%s ...", canvasID)

	// locate the canvas
	document := js.Global().Get("document")
	canvasElement := document.Call("getElementById", canvasID)
	screen := newScreen(canvasElement)
	log.Printf("canvas dimensions device ratio:%f width:%d x height:%d",
		screen.devicePixelRatio, screen.devicePixelWidth, screen.devicePixelHeight)

	g := game.New()
	c := newClient(screen, g)
	defer c.Stop()

	document.Call("addEventListener", "keydown", c.keyDownCallback)
	document.Call("addEventListener", "keyup", c.keyUpCallback)

	js.Global().Call("requestAnimationFrame", c.requestFrame)

	done := make(chan struct{})
	<-done
}

func newScreen(canvasElement js.Value) *canvasScreen {
	// make the canvas use REAL pixels so we can draw the real pixels without scaling
	// https://stackoverflow.com/a/59511599
	devicePixelRatio := js.Global().Get("devicePixelRatio").Float()
	width := int(canvasElement.Get("width").Float())
	height := int(canvasElement.Get("height").Float())
	if devicePixelRatio != 1.0 {
		log.Printf("devicePixelRatio=%f", devicePixelRatio)

		canvasStyle := canvasElement.Get("style")
		canvasStyle.Set("width", fmt.Sprintf("%dpx", width))
		canvasStyle.Set("height", fmt.Sprintf("%dpx", height))

		// now make the canvas bigger
		width = int(float64(width) * devicePixelRatio)
		height = int(float64(height) * devicePixelRatio)
		canvasElement.Set("width", width)
		canvasElement.Set("height", height)
	}

	// Create a Go image to draw into
	frame := image.NewRGBA(image.Rect(0, 0, width, height))
	gc := draw2dimg.NewGraphicContext(frame)
	gc.Scale(devicePixelRatio, devicePixelRatio)

	// set up the Canvas so we can draw to it
	ctx := canvasElement.Call("getContext", "2d")
	imageData := ctx.Call("createImageData", js.ValueOf(width), js.ValueOf(height))
	imageDataData := imageData.Get("data")
	jsUInt8Array := js.Global().Get("Uint8Array").New(len(frame.Pix))

	return &canvasScreen{devicePixelRatio, width, height, ctx, imageData, imageDataData, jsUInt8Array, frame, gc}
	// c.image = image.NewRGBA(image.Rect(0, 0, width, height))
	// c.copybuff = js.Global().Get("Uint8Array").New(len(c.image.Pix)) // Static JS buffer for copying data out to JS. Defined once and re-used to save on un-needed allocations
}

type canvasScreen struct {
	devicePixelRatio  float64
	devicePixelWidth  int
	devicePixelHeight int

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
