package main

import (
	"fmt"
	"image"
	"log"
	"syscall/js"

	"github.com/evanj/netgamesim/intersect"
	"github.com/evanj/netgamesim/sprites"
	"github.com/llgcode/draw2d/draw2dimg"
)

const canvasID = "canvas"

// we simulate at a fixed timestep 16 ms = 62.5 FPS which is really close to the 60 FPS
// target for web games
// see https://gafferongames.com/post/fix_your_timestep/
const simulationTimeStepMS = 16

const tankMovePerSecond = 300.0
const tankMovePerTimeStep = tankMovePerSecond / 1000.0 * simulationTimeStepMS

const targetMovePerSecond = 400.0
const targetMovePerTimeStep = targetMovePerSecond / 1000.0 * simulationTimeStepMS

const bulletMovePerSecond = 900.0
const bulletMovePerTimeStep = bulletMovePerSecond / 1000.0 * simulationTimeStepMS

const smokeDisplaySeconds = 1
const smokeDisplayTimeSteps = (smokeDisplaySeconds * 1000) / simulationTimeStepMS

const tankInitialX = 75
const tankInitialY = 75

const targetX = 400
const targetMaxY = 450
const targetMinY = 50

const keyCodeSpace = 32
const keyCodeLeft = 37
const keyCodeUp = 38
const keyCodeRight = 39
const keyCodeDown = 40

const logFPSSeconds = 15

const maxEdgeX = 500

type direction int

const (
	dirNone = direction(iota)
	dirLeft
	dirUp
	dirRight
	dirDown
)

type smokeCounter struct {
	position      intersect.Point
	timeStepCount int
}

type game struct {
	sprites         sprites.Sprites
	keyDownCallback js.Func
	keyUpCallback   js.Func
	requestFrame    js.Func
	screen          *canvasScreen
	simTime         float64
	lastFPSLogTime  float64
	frames          int

	tank      intersect.Point
	dir       direction
	target    intersect.Point
	targetDir direction
	bullets   []intersect.Point
	smoke     []smokeCounter

	firePressed bool
}

func newGame(screen *canvasScreen) *game {
	g := &game{sprites.New(), js.Func{}, js.Func{}, js.Func{}, screen, 0.0, 0.0, 0,
		intersect.Point{X: tankInitialX, Y: tankInitialY}, dirNone,
		intersect.Point{X: targetX, Y: targetMinY}, dirDown,
		nil, nil, false}
	g.keyDownCallback = js.FuncOf(g.jsKeyDown)
	g.keyUpCallback = js.FuncOf(g.jsKeyUp)
	g.requestFrame = js.FuncOf(g.jsRequestFrame)
	return g
}

func (g *game) Stop() {
	g.keyDownCallback.Release()
	g.keyUpCallback.Release()
	g.requestFrame.Release()
}

func dirFromKeyCode(keyCode int) direction {
	switch keyCode {
	case keyCodeLeft:
		return dirLeft
	case keyCodeUp:
		return dirUp
	case keyCodeRight:
		return dirRight
	case keyCodeDown:
		return dirDown
	default:
		return dirNone
	}
}

func (g *game) jsKeyDown(this js.Value, args []js.Value) interface{} {
	event := args[0]
	// TODO: check for repeat?
	// repeat := event.Get("repeat").Bool()
	// probably not worth it if this involves a "call" back to the browser?

	keyCode := event.Get("keyCode").Int()
	if keyCode == keyCodeSpace {
		if g.firePressed {
			// ignore duplicate
			return nil
		}
		g.firePressed = true
		g.bullets = append(g.bullets, g.tank)
	} else {
		dir := dirFromKeyCode(keyCode)

		if dir == dirNone {
			return nil
		}
		g.dir = dir
	}

	// prevent arrow keys from doing what they normally would
	event.Call("preventDefault")
	return nil
}

func (g *game) jsKeyUp(this js.Value, args []js.Value) interface{} {
	event := args[0]
	keyCode := event.Get("keyCode").Int()

	if keyCode == keyCodeSpace {
		g.firePressed = false
		return nil
	}

	dir := dirFromKeyCode(keyCode)
	if dir == dirNone {
		return nil
	}
	if dir == g.dir {
		g.dir = dirNone
	}
	return nil
}

func (g *game) simulateTimeStep() {
	offsetX := 0.0
	offsetY := 0.0

	switch g.dir {
	case dirLeft:
		offsetX = -tankMovePerTimeStep
	case dirRight:
		offsetX = tankMovePerTimeStep

	case dirDown:
		offsetY = tankMovePerTimeStep
	case dirUp:
		offsetY = -tankMovePerTimeStep

	case dirNone:
		// do nothing

	default:
		panic("unhandled direction")
	}

	g.tank.X += offsetX
	g.tank.Y += offsetY

	switch g.targetDir {
	case dirDown:
		g.target.Y += targetMovePerTimeStep
		if g.target.Y > targetMaxY {
			g.target.Y = targetMaxY
			g.targetDir = dirUp
		}
	case dirUp:
		g.target.Y -= targetMovePerTimeStep
		if g.target.Y < targetMinY {
			g.target.Y = targetMinY
			g.targetDir = dirDown
		}
	default:
		panic("bad target direction")
	}

	for i := 0; i < len(g.bullets); i++ {
		g.bullets[i].X += bulletMovePerTimeStep

		shouldRemove := false
		if g.bullets[i].X >= maxEdgeX {
			// bullet is off the screen: remove it
			shouldRemove = true
		}

		// in testing: the point/box intersection is basically as good as the the path/box
		// intersection and much simpler. It misses on RARE occasions
		if intersect.PointBox(g.bullets[i], g.target, g.sprites.Target.Size) {
			// bullet hit the target! remove it and add smoke
			shouldRemove = true
			g.smoke = append(g.smoke, smokeCounter{g.bullets[i], 0})
			log.Printf("hit! bullet = %s ; target = %s", g.bullets[i], g.target)
		}

		if shouldRemove {
			last := len(g.bullets) - 1
			g.bullets[last], g.bullets[i] = g.bullets[i], g.bullets[last]
			g.bullets = g.bullets[:last]
			i--
		}
	}

	for i := 0; i < len(g.smoke); i++ {
		g.smoke[i].timeStepCount++
		if g.smoke[i].timeStepCount >= smokeDisplayTimeSteps {
			last := len(g.smoke) - 1
			g.smoke[last], g.smoke[i] = g.smoke[i], g.smoke[last]
			g.smoke = g.smoke[:last]
			i--
		}
	}
}

func (g *game) jsRequestFrame(this js.Value, args []js.Value) interface{} {
	msSinceDocStart := args[0].Float()
	if g.simTime == 0.0 {
		g.simTime = msSinceDocStart
		g.lastFPSLogTime = msSinceDocStart
	}

	// advance physics simulation until we are "caught up"
	// see https://gafferongames.com/post/fix_your_timestep/
	frames := 0
	for {
		nextTime := g.simTime + simulationTimeStepMS
		if nextTime >= msSinceDocStart {
			break
		}
		g.simTime = nextTime

		g.simulateTimeStep()
		frames++
	}

	// draw the state of the universe
	g.sprites.Tank.Draw(g.screen.gc, g.tank.X, g.tank.Y)
	g.sprites.Target.Draw(g.screen.gc, g.target.X, g.target.Y)
	for _, b := range g.bullets {
		g.sprites.Bullet.Draw(g.screen.gc, b.X, b.Y)
	}
	for _, s := range g.smoke {
		g.sprites.Smoke.Draw(g.screen.gc, s.position.X, s.position.Y)
	}
	g.screen.renderFrame()

	// request the next frame
	js.Global().Call("requestAnimationFrame", g.requestFrame)

	g.frames++
	if msSinceDocStart-g.lastFPSLogTime >= logFPSSeconds*1000 {
		seconds := (msSinceDocStart - g.lastFPSLogTime) / 1000.0
		fps := float64(g.frames) / seconds
		log.Printf("t=%f frames=%d seconds=%f fps=%f", msSinceDocStart, g.frames, seconds, fps)
		g.frames = 0
		g.lastFPSLogTime = msSinceDocStart
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

	g := newGame(screen)
	defer g.Stop()

	document.Call("addEventListener", "keydown", g.keyDownCallback)
	document.Call("addEventListener", "keyup", g.keyUpCallback)

	js.Global().Call("requestAnimationFrame", g.requestFrame)

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
