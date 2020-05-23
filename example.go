package main

import (
	"errors"
	// "fmt"

	"github.com/gopherjs/gopherjs/js"
)

type KeyCode int

const (
	left  KeyCode = 37
	up            = 38
	right         = 39
	down          = 40

	space = 32
)

func KeyCodeFromInt(v int) (error, KeyCode) {
	if (int(left) <= v && v <= int(down)) || v == space {
		return nil, KeyCode(v)
	}
	// return fmt.Errorf("unsupported key code: %d", v), KeyCode(0)
	return errors.New("unsupported key code"), KeyCode(0)
}

type Player struct {
	directionX int
	directionY int
	tank       *Tank
}

func (p *Player) Direction(x int, y int) {
	p.directionX = x
	p.directionY = y
}

type World struct {
	player  *Player
	bullets []point
}

func (w *World) addBullet(x int, y int) {
	w.bullets = append(w.bullets, point{x, y})
}

func (w *World) simulate(milliseconds float64) {
	// TODO: Normalize based on time!
	// fmt.Println("simulate", seconds)
	if w.player.directionX != 0 {
		w.player.tank.x += w.player.directionX
	}
	if w.player.directionY != 0 {
		w.player.tank.y += w.player.directionY
	}
}

func (w *World) Draw(canvas *Canvas) {
	w.player.tank.Draw(canvas)

	for _, p := range w.bullets {
		canvas.line(p.x, p.y, p.x, p.y+4)
	}
}

func (w *World) KeyDown(keycode KeyCode) {
	switch keycode {
	case left:
		w.player.directionX = -1
	case right:
		w.player.directionX = 1
	case up:
		w.player.directionY = -1
	case down:
		w.player.directionY = 1
	case space:
		w.addBullet(w.player.tank.x, w.player.tank.y+tankSize)
	default:
		panic("unsupported keycode")
	}
	// fmt.Println("keydown", w.player.directionX, w.player.directionY)
}

func (w *World) KeyUp(keycode KeyCode) {
	switch keycode {
	case left, right:
		w.player.directionX = 0
	case up, down:
		w.player.directionY = 0
	case space:
		// do nothing
	default:
		panic("unsupported keycode")
	}
	// fmt.Println("keyup", w.player.directionX, w.player.directionY)
}

type point struct {
	x int
	y int
}

type Tank struct {
	x     int
	y     int
	angle float64
}

const tankSize = 10

func (t *Tank) Draw(canvas *Canvas) {
	canvas.strokeRect(t.x-tankSize/2, t.y-tankSize/2, tankSize, tankSize)
	canvas.line(t.x, t.y, t.x, t.y+tankSize)
}

func getEventKey(event *js.Object) (error, KeyCode) {
	return KeyCodeFromInt(event.Get("keyCode").Int())
}

type Canvas struct {
	jsContext      *js.Object
	cssPixelWidth  float64
	cssPixelOffset float64

	cssWidth  int
	cssHeight int
}

func (c *Canvas) pixPos(v int) float64 {
	return c.pixLen(v) + c.cssPixelOffset
}

func (c *Canvas) pixLen(v int) float64 {
	return float64(v) * c.cssPixelWidth
}

func (c *Canvas) clear() {
	width := float64(c.cssWidth) * c.cssPixelWidth
	height := float64(c.cssHeight) * c.cssPixelWidth
	c.jsContext.Call("clearRect", 0.0, 0.0, width, height)
}

func (c *Canvas) strokeRect(x int, y int, width int, height int) {
	c.jsContext.Call("strokeRect", c.pixPos(x), c.pixPos(y), c.pixLen(width), c.pixLen(height))
}

func (c *Canvas) fillRect(x int, y int, width int, height int) {
	c.jsContext.Call("fillRect", c.pixPos(x), c.pixPos(y), c.pixLen(width), c.pixLen(height))
}

func (c *Canvas) line(x1 int, y1 int, x2 int, y2 int) {
	c.jsContext.Call("beginPath")
	c.jsContext.Call("moveTo", c.pixPos(x1), c.pixPos(y1))
	c.jsContext.Call("lineTo", c.pixPos(x2), c.pixPos(y2))
	c.jsContext.Call("stroke")
}

func JSNewGame(jsContext *js.Object) *js.Object {
	pixelWidth := js.Global.Get("devicePixelRatio").Float()
	canvas := &Canvas{jsContext, pixelWidth, pixelWidth / 2, 500, 500}
	jsContext.Set("lineWidth", canvas.cssPixelWidth)

	world := &World{}
	world.player = &Player{0, 0, &Tank{100, 100, 0}}

	jsg := &JSGame{world, canvas, 0}
	jsg.requestFrame()
	return js.MakeWrapper(jsg)
}

type JSGame struct {
	world     *World
	canvas    *Canvas
	lastFrame float64
}

func (jsg *JSGame) KeyDown(event *js.Object) {
	// ignore unsupported keys
	err, keycode := getEventKey(event)
	if err != nil {
		return
	}
	// drop repeates
	repeat := event.Get("repeat").Bool()
	if repeat {
		return
	}

	jsg.world.KeyDown(keycode)
	event.Call("preventDefault")
}

func (jsg *JSGame) KeyUp(event *js.Object) {
	// ignore unsupported keys
	err, keycode := getEventKey(event)
	if err != nil {
		return
	}

	jsg.world.KeyUp(keycode)
}

func (jsg *JSGame) Frame(domHighResTimeStamp *js.Object) {
	timestamp := domHighResTimeStamp.Float()
	if jsg.lastFrame == 0 {
		jsg.lastFrame = timestamp
		jsg.requestFrame()
		return
	}
	offset := timestamp - jsg.lastFrame
	// fmt.Printf("frame timestamp: %f offset: %f\n", timestamp, offset)

	jsg.canvas.clear()
	jsg.world.simulate(offset)
	jsg.world.Draw(jsg.canvas)
	jsg.requestFrame()
}

func (jsg *JSGame) requestFrame() {
	// TODO: Cache the conversion of the method to JS argument?
	js.Global.Call("requestAnimationFrame", jsg.Frame)
}

func main() {
	js.Global.Set("dumbgame", map[string]interface{}{
		"NewGame": JSNewGame,
	})
}
