//go:build wasm
// +build wasm

package main

import (
	"fmt"
	"image"
	"log"
	"math"
	"syscall/js"

	"github.com/evanj/netgamesim/game"
	"github.com/evanj/netgamesim/sprites"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

const clientCanvasID = "clientCanvas"
const serverCanvasID = "serverCanvas"

const keyCodeSpace = 32
const keyCodeLeft = 37
const keyCodeUp = 38
const keyCodeRight = 39
const keyCodeDown = 40

const logFPSSeconds = 15

const tapMS = 100
const touchMovePixels = 30

type client struct {
	keyDownCallback    js.Func
	keyUpCallback      js.Func
	touchStartCallback js.Func
	touchMoveCallback  js.Func
	touchEndCallback   js.Func

	game *game.Game

	fireKeyDown bool
	sendFire    bool
	tankDir     game.Direction

	touchStartMS float64
	touchX       float64
	touchY       float64
}

func newClient(g *game.Game) *client {
	c := &client{
		js.Func{}, js.Func{}, js.Func{}, js.Func{}, js.Func{},
		g,
		false, false, game.DirNone,
		0.0, 0.0, 0.0,
	}
	c.keyDownCallback = js.FuncOf(c.jsKeyDown)
	c.keyUpCallback = js.FuncOf(c.jsKeyUp)
	c.touchStartCallback = js.FuncOf(c.jsTouchStart)
	c.touchMoveCallback = js.FuncOf(c.jsTouchMove)
	c.touchEndCallback = js.FuncOf(c.jsTouchEnd)
	return c
}

func (c *client) Stop() {
	c.keyDownCallback.Release()
	c.keyUpCallback.Release()
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

	keyCode := event.Get("keyCode").Int()
	switch keyCode {
	case keyCodeSpace:
		if c.fireKeyDown {
			// ignore duplicate
			event.Call("preventDefault")
			return nil
		}
		c.fireKeyDown = true
		c.sendFire = true

	case keyCodeLeft, keyCodeUp, keyCodeRight, keyCodeDown:
		dir := dirFromKeyCode(keyCode)
		if dir == game.DirNone {
			panic("BUG: mismatch between case and dirFromKeyCode")
		}
		c.tankDir = dir

	default:
		// unknown key: ignore
		return nil
	}

	// prevent keys from doing what they normally would
	event.Call("preventDefault")
	return nil
}

func (c *client) jsKeyUp(this js.Value, args []js.Value) interface{} {
	event := args[0]
	keyCode := event.Get("keyCode").Int()

	if keyCode == keyCodeSpace {
		c.fireKeyDown = false
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

func (c *client) jsTouchStart(this js.Value, args []js.Value) interface{} {
	event := args[0]
	touches := event.Get("touches")
	touches0 := touches.Get("0")
	c.touchStartMS = event.Get("timeStamp").Float()
	c.touchX = touches0.Get("pageX").Float()
	c.touchY = touches0.Get("pageY").Float()
	log.Printf("touch start x:%f y:%f", c.touchX, c.touchY)

	event.Call("preventDefault")
	return nil
}

func (c *client) jsTouchMove(this js.Value, args []js.Value) interface{} {
	// log.Printf("touch move")
	event := args[0]
	touches := event.Get("touches")
	touches0 := touches.Get("0")
	x := touches0.Get("pageX").Float()
	y := touches0.Get("pageY").Float()
	xDiff := x - c.touchX
	yDiff := y - c.touchY

	xDiffAbs := math.Abs(xDiff)
	yDiffAbs := math.Abs(yDiff)
	logMove := false
	if xDiffAbs > yDiffAbs && xDiffAbs > touchMovePixels {
		if c.tankDir == game.DirNone {
			logMove = true
		}
		if xDiff < 0 {
			// joystick move left
			c.tankDir = game.DirLeft
		} else {
			c.tankDir = game.DirRight
		}
	} else if yDiffAbs > touchMovePixels {
		if c.tankDir == game.DirNone {
			logMove = true
		}
		if yDiff < 0 {
			// joystick move up
			c.tankDir = game.DirUp
		} else {
			c.tankDir = game.DirDown
		}
	}
	if logMove {
		log.Printf("touch joystick move xDiff:%f yDiff:%f", xDiff, yDiff)
	}

	event.Call("preventDefault")
	return nil
}

func (c *client) jsTouchEnd(this js.Value, args []js.Value) interface{} {
	event := args[0]

	if c.tankDir != game.DirNone {
		// this was a move! cancel it
		c.tankDir = game.DirNone
	} else {
		// check for tap
		time := event.Get("timeStamp").Float()
		if time-c.touchStartMS <= tapMS {
			// this is a tap! fire
			c.sendFire = true
		}
		log.Printf("tap time = %f", time-c.touchStartMS)
	}

	event.Call("preventDefault")
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

type clientMessage struct {
	sentTime float64
	input    game.Input
}

type serverMessage struct {
	sentTime float64
	state    *game.Game
}

type network struct {
	currentMS float64
	latencyMS float64

	clientToServer []clientMessage
	serverToClient []serverMessage
}

func (n *network) getServerIncoming(current float64) *game.Input {
	if len(n.clientToServer) == 0 {
		return nil
	}

	// all messages before this time should be delivered
	deliveredSentTime := current - n.latencyMS
	if n.clientToServer[0].sentTime <= deliveredSentTime {
		out := n.clientToServer[0].input
		n.clientToServer = n.clientToServer[1:]
		return &out
	}
	return nil
}

func (n *network) getClientIncoming(current float64) *game.Game {
	if len(n.serverToClient) == 0 {
		return nil
	}

	// all messages before this time should be delivered
	deliveredSentTime := current - n.latencyMS
	if n.serverToClient[0].sentTime <= deliveredSentTime {
		out := n.serverToClient[0].state
		n.serverToClient = n.serverToClient[1:]
		return out
	}
	return nil
}

func (n *network) sendToClient(current float64, g *game.Game) {
	n.serverToClient = append(n.serverToClient, serverMessage{current, g})
}

func (n *network) sendToServer(current float64, i game.Input) {
	n.clientToServer = append(n.clientToServer, clientMessage{current, i})
}

type server struct {
	game *game.Game
}

func newServer() *server {
	return &server{game.New()}
}

func (s *server) executeTimeStep() *game.Game {
	s.game.SimulateTimeStep()
	return s.game.Clone()
}

type simulation struct {
	simTimeStart float64
	net          *network

	client       *client
	clientScreen *canvasScreen

	server       *server
	serverScreen *canvasScreen

	requestFrame    js.Func
	latencyAdjusted js.Func

	lastFPSLogTime float64
	frames         int
}

func newSimulation(clientScreen *canvasScreen, serverScreen *canvasScreen) *simulation {
	sim := &simulation{
		0.0, &network{},

		newClient(game.New()), clientScreen,

		newServer(), serverScreen,

		js.Func{}, js.Func{},

		0.0, 0,
	}
	sim.requestFrame = js.FuncOf(sim.jsRequestFrame)
	sim.latencyAdjusted = js.FuncOf(sim.jsLatencyAdjusted)
	return sim
}

func (s *simulation) Stop() {
	s.latencyAdjusted.Release()
	s.client.Stop()
}

func (s *simulation) jsRequestFrame(this js.Value, args []js.Value) interface{} {
	msSinceDocStart := args[0].Float()
	if s.simTimeStart == 0.0 {
		s.simTimeStart = msSinceDocStart
		s.lastFPSLogTime = msSinceDocStart
	}

	msSinceStart := msSinceDocStart - s.simTimeStart

	// simulate the network advancing by single ticks; we can't show anything more often than 60
	// fps anyway, so latency is "quantized" to frames anaway
	for serverTime := s.net.currentMS + game.TimeStepMS; serverTime < msSinceStart; serverTime += game.TimeStepMS {
		// process server network input
		for {
			input := s.net.getServerIncoming(serverTime)
			if input == nil {
				break
			}
			s.server.game.ProcessInput(*input)
		}

		// simulate the time on the server; send the updated state to the client
		state := s.server.executeTimeStep()
		s.net.sendToClient(serverTime, state)

		// process client network messages by replacing the game state
		for {
			state := s.net.getClientIncoming(serverTime)
			if state == nil {
				break
			}
			s.client.game = state
		}

		s.net.currentMS = serverTime
	}
	// client sends a message to the server every frame
	input := game.Input{
		TankDir: s.client.tankDir,
		Fire:    s.client.sendFire,
	}
	s.client.sendFire = false
	s.net.sendToServer(msSinceStart, input)

	// draw the state of the universe
	drawGame(s.clientScreen.gc, s.client.game)
	s.clientScreen.renderFrame()
	drawGame(s.serverScreen.gc, s.server.game)
	s.serverScreen.renderFrame()

	// request the next frame
	js.Global().Call("requestAnimationFrame", s.requestFrame)

	s.frames++
	if msSinceDocStart-s.lastFPSLogTime >= logFPSSeconds*1000 {
		seconds := (msSinceDocStart - s.lastFPSLogTime) / 1000.0
		fps := float64(s.frames) / seconds
		log.Printf("t=%f frames=%d seconds=%f fps=%f", msSinceDocStart, s.frames, seconds, fps)
		s.frames = 0
		s.lastFPSLogTime = msSinceDocStart
	}
	return nil
}

func (s *simulation) jsLatencyAdjusted(this js.Value, args []js.Value) interface{} {
	v := args[0].Float()
	log.Printf("latency adjusted = %f", v)
	s.net.latencyMS = v
	return nil
}

func main() {
	log.Printf("demo loading in client canvas=%s; server canvas=%s ...",
		clientCanvasID, serverCanvasID)

	// locate the client canvas
	document := js.Global().Get("document")
	clientCanvasElement := document.Call("getElementById", clientCanvasID)
	clientScreen := newScreen(clientCanvasElement)
	log.Printf("client canvas dimensions device ratio:%f width:%d x height:%d",
		clientScreen.devicePixelRatio, clientScreen.devicePixelWidth, clientScreen.devicePixelHeight)

	serverCanvasElement := document.Call("getElementById", serverCanvasID)
	serverScreen := newScreen(serverCanvasElement)

	s := newSimulation(clientScreen, serverScreen)
	defer s.Stop()

	document.Call("addEventListener", "keydown", s.client.keyDownCallback)
	document.Call("addEventListener", "keyup", s.client.keyUpCallback)
	clientCanvasElement.Call("addEventListener", "touchstart", s.client.touchStartCallback)
	clientCanvasElement.Call("addEventListener", "touchend", s.client.touchEndCallback)
	clientCanvasElement.Call("addEventListener", "touchcancel", s.client.touchEndCallback)
	clientCanvasElement.Call("addEventListener", "touchmove", s.client.touchMoveCallback)

	js.Global().Call("requestAnimationFrame", s.requestFrame)
	js.Global().Set("gameLatencyAdjusted", s.latencyAdjusted)

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
