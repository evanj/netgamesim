package game

import (
	"log"

	"github.com/evanj/netgamesim/intersect"
	"github.com/evanj/netgamesim/sprites"
)

// the game world is 500x500
const maxEdgeDimension = 500

// we simulate at a fixed timestep 16 ms = 62.5 FPS which is really close to the 60 FPS
// target for web games
// see https://gafferongames.com/post/fix_your_timestep/
const TimeStepMS = 16

const tankMovePerSecond = 300
const tankMovePerTimeStep = (tankMovePerSecond * TimeStepMS) / 1000.0

const targetMovePerSecond = 400
const targetMovePerTimeStep = (targetMovePerSecond * TimeStepMS) / 1000.0

const bulletMovePerSecond = 900
const bulletMovePerTimeStep = (bulletMovePerSecond * TimeStepMS) / 1000.0

const smokeDisplaySeconds = 1
const smokeDisplayTimeSteps = int((smokeDisplaySeconds*1000.0)/TimeStepMS + 0.5)

const tankInitialX = 75
const tankInitialY = 75

const targetX = 400
const targetMaxY = 450
const targetMinY = 50

type Direction int

const (
	DirNone = Direction(iota)
	DirLeft
	DirUp
	DirRight
	DirDown
)

type Event int

const (
	EventNone = Event(iota)
	EventFire
)

type smoke struct {
	position      intersect.Point
	timeStepCount int
}

// Game contains the state of the world and can advance the simulation.
// It does not know how to move
type Game struct {
	tank    intersect.Point
	tankDir Direction

	target    intersect.Point
	targetDir Direction

	bullets []intersect.Point
	smoke   []smoke

	simTicks int
}

// TankCenter returns the current tank center.
func (g *Game) TankCenter() intersect.Point { return g.tank }

// TargetCenter returns the current target center.
func (g *Game) TargetCenter() intersect.Point { return g.target }

// Bullets returns the current bullet locations.
func (g *Game) Bullets() []intersect.Point { return g.bullets }

// Smoke returns the current smoke locations.
func (g *Game) Smoke() []intersect.Point {
	// TODO: This is an inefficient allocation/copy; fix?
	p := make([]intersect.Point, len(g.smoke))
	for i, s := range g.smoke {
		p[i] = s.position
	}
	return p
}

func New() *Game {
	g := &Game{
		// tank
		intersect.Point{X: tankInitialX, Y: tankInitialY}, DirNone,
		// target
		intersect.Point{X: targetX, Y: targetMinY}, DirDown,
		nil, nil,
		0,
	}
	return g
}

func (g *Game) Clone() *Game {
	bulletsClone := make([]intersect.Point, len(g.bullets))
	for i, b := range g.bullets {
		bulletsClone[i] = b
	}
	smokeClone := make([]smoke, len(g.smoke))
	for i, s := range g.smoke {
		smokeClone[i] = s
	}
	return &Game{
		g.tank, g.tankDir, g.target, g.targetDir, bulletsClone, smokeClone, g.simTicks,
	}
}

type Input struct {
	TankDir Direction
	Fire    bool
}

// ProcessInput processes the input from the player.
func (g *Game) ProcessInput(i Input) {
	g.tankDir = i.TankDir

	if i.Fire {
		g.bullets = append(g.bullets, g.tank)
	}
}

// AdvanceSimulation advances the simulation to msSinceStart.
func (g *Game) AdvanceSimulation(msSinceStart float64) {
	ticksSinceStart := int(msSinceStart/TimeStepMS)

	// advance physics simulation until we are "caught up"
	// see https://gafferongames.com/post/fix_your_timestep/
	for g.simTicks < ticksSinceStart {
		g.SimulateTimeStep()
	}
}

func (g *Game) SimulateTimeStep() {
	offsetX := 0.0
	offsetY := 0.0

	switch g.tankDir {
	case DirLeft:
		offsetX = -tankMovePerTimeStep
	case DirRight:
		offsetX = tankMovePerTimeStep

	case DirDown:
		offsetY = tankMovePerTimeStep
	case DirUp:
		offsetY = -tankMovePerTimeStep

	case DirNone:
		// do nothing

	default:
		panic("unhandled direction")
	}

	g.tank.X += offsetX
	g.tank.Y += offsetY

	switch g.targetDir {
	case DirDown:
		g.target.Y += targetMovePerTimeStep
		if g.target.Y > targetMaxY {
			g.target.Y = targetMaxY
			g.targetDir = DirUp
		}
	case DirUp:
		g.target.Y -= targetMovePerTimeStep
		if g.target.Y < targetMinY {
			g.target.Y = targetMinY
			g.targetDir = DirDown
		}
	default:
		panic("bad target direction")
	}

	for i := 0; i < len(g.bullets); i++ {
		g.bullets[i].X += bulletMovePerTimeStep

		shouldRemove := false
		if g.bullets[i].X >= maxEdgeDimension {
			// bullet is off the screen: remove it
			shouldRemove = true
		}

		// in testing: the point/box intersection is basically as good as the the path/box
		// intersection and much simpler. It misses on RARE occasions
		if intersect.PointBox(g.bullets[i], g.target, sprites.TargetSize) {
			// bullet hit the target! remove it and add smoke
			shouldRemove = true
			g.smoke = append(g.smoke, smoke{g.bullets[i], 0})
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

	g.simTicks++
}
