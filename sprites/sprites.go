package sprites

import (
	"image"
	"image/color"
	"log"
	"math"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
)

var tankGreen = color.RGBA{0x0c, 0xd4, 0x63, 0xff}
var targetDarkRed = color.RGBA{0x9a, 0x1f, 0x40, 0xff}
var targetLightRed = color.RGBA{0xd9, 0x45, 0x5f, 0xff}
var smokeGrey = color.RGBA{0xaa, 0xaa, 0xaa, 0xaa}

const tankSize = 20
const tankLineWidth = 2.0
// const TankCenterX = tankSize / 2
// const TankCenterY = tankSize / 2

const targetSize = 15
const targetInnerSize = targetSize/2.0

const bulletSize = 4
const smokeSize = 25

const pixelStrokeOffset = 0.5

// Sprites are assumed to be squares Size x Size in dimension. It is actually
// okay to draw the sprite past the Size boundary.
type Sprite struct {
	Size float64
	Draw func(gc draw2d.GraphicContext, centerX float64, centerY float64)
}

type Sprites struct {
	Tank Sprite
	Target Sprite
	Bullet Sprite
	Smoke Sprite
}

func New() Sprites {
	return Sprites{
		Sprite{tankSize, drawTank},
		Sprite{targetSize, drawTarget},
		Sprite{bulletSize, drawBullet},
		Sprite{smokeSize, drawSmoke},
	}
}

func drawTank(gc draw2d.GraphicContext, centerX float64, centerY float64) {
	gc.SetStrokeColor(color.Black)
	gc.SetFillColor(tankGreen)
	gc.SetLineWidth(tankLineWidth)
	// this should make the corners of the tank square but it has a bug:
	// https://github.com/llgcode/draw2d/issues/155
	// sadly this means we need to draw some insane lines with rectangles
	gc.SetLineJoin(draw2d.MiterJoin)

	// TODO: should this be for all odd line widths? E.g. 3.0, 5.0? what about 3.5 or 0.5?
	if tankLineWidth <= 1.0 {
		centerX += pixelStrokeOffset
		centerY += pixelStrokeOffset
	}

	// tank "body"
	gc.MoveTo(centerX-tankSize/2, centerY-tankSize/2)
	gc.LineTo(centerX+tankSize/2, centerY-tankSize/2)
	gc.LineTo(centerX+tankSize/2, centerY+tankSize/2)
	gc.LineTo(centerX-tankSize/2, centerY+tankSize/2)
	gc.Close()

	// tank "gun"
	gc.MoveTo(centerX, centerY)
	gc.LineTo(centerX+tankSize, centerY)
	gc.FillStroke()

	gc.BeginPath()
}

func drawTarget(gc draw2d.GraphicContext, centerX float64, centerY float64) {
	gc.SetFillColor(targetDarkRed)
	draw2dkit.Circle(gc, centerX, centerY, targetSize)
	gc.Fill()
	gc.SetFillColor(targetLightRed)
	draw2dkit.Circle(gc, centerX, centerY, targetInnerSize)
	gc.Fill()
}

func drawBullet(gc draw2d.GraphicContext, centerX float64, centerY float64) {
	gc.SetFillColor(color.Black)
	draw2dkit.Circle(gc, centerX, centerY, bulletSize)
	gc.Fill()
}

func drawSmoke(gc draw2d.GraphicContext, centerX float64, centerY float64) {
	gc.SetFillColor(smokeGrey)
	draw2dkit.Circle(gc, centerX, centerY, smokeSize)
	gc.Fill()
}

// fillPixels fills the rectangle described by image.Rectangle with the given color
func fillPixels(gc draw2d.GraphicContext, rect image.Rectangle, c color.Color) {
	gc.SetFillColor(c)
	pathRectangle(gc, rect, 0.0)
	gc.Fill()
}

// draw1Px draws a single pixel think line around the inclusive pixels described by rect.
// For example, (1,1) -> (4,4) is a 4 pixel wide rectangle that includes pixels (1,1) and (4,4)
func draw1PxRect(gc draw2d.GraphicContext, rect image.Rectangle, c color.Color) {
	gc.SetStrokeColor(c)
	pathRectangle(gc, rect, pixelStrokeOffset)
	gc.Stroke()
}

func iabs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// draw1Px draws a single pixel thick line that includes p1 and p2. That is, a line
// from (1,1) -> (1,4) is a 4 pixels vertically and 1 pixel horizontally
func draw1PxLine(gc draw2d.GraphicContext, p1 image.Point, p2 image.Point, c color.Color) {
	// determine the "primary" direction as the largest magnitude
	xDiff := p2.X - p1.X
	yDiff := p2.Y - p1.Y

	xStartOffset := pixelStrokeOffset
	xEndOffset := pixelStrokeOffset
	yStartOffset := pixelStrokeOffset
	yEndOffset := pixelStrokeOffset
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

	log.Printf("draw1PxLine (%d,%d)->(%d,%d) === (%f,%f)->(%f,%f)",
		p1.X, p1.Y, p2.X, p2.Y,
		float64(p1.X)+xStartOffset, float64(p1.Y)+yStartOffset, float64(p2.X)+xEndOffset, float64(p2.Y)+yEndOffset)
}

func pathRectangle(gc draw2d.GraphicContext, rect image.Rectangle, offset float64) {
	gc.BeginPath()
	gc.MoveTo(float64(rect.Min.X)+offset, float64(rect.Min.Y)+offset)
	gc.LineTo(float64(rect.Max.X)+offset, float64(rect.Min.Y)+offset)
	gc.LineTo(float64(rect.Max.X)+offset, float64(rect.Max.Y)+offset)
	gc.LineTo(float64(rect.Min.X)+offset, float64(rect.Max.Y)+offset)
	gc.Close()
}

func angles() *image.RGBA {
	const radius = 10
	const spacing = 1
	const divisions = 32
	const totalRads = 2 * math.Pi

	img := image.NewRGBA(image.Rect(0, 0, (2*radius+spacing)*divisions, 2*radius+2*spacing))
	gc := draw2dimg.NewGraphicContext(img)

	for div := 0; div < divisions; div += 1 {
		startX := spacing*div + (2 * radius * div) + radius
		startY := spacing + radius

		rads := (totalRads / float64(divisions)) * float64(div)
		x := math.Cos(rads) * radius
		y := math.Sin(rads) * radius

		xInt := startX + int(x+0.5)
		yInt := startY + int(y+0.5)
		log.Printf("division %d = rads %f; --> %f,%f = (%d,%d) line (%d,%d)", div, rads, x, y, startX, startY, xInt, yInt)

		draw1PxLine(gc, image.Point{startX, startY}, image.Point{xInt, yInt}, color.Black)
		// gc.SetStrokeColor(color.Black)
		// gc.SetLineWidth(1.0)
		// gc.BeginPath()
		// gc.MoveTo(float64(startX), float64(startY))
		// gc.LineTo(float64(xInt), float64(yInt))
		// gc.Stroke()
	}

	return img
}

func angles2() *image.RGBA {
	const radius = 100
	const spacing = 1
	const divisions = 32
	const totalRads = 2 * math.Pi

	img := image.NewRGBA(image.Rect(0, 0, (2*radius + 2*spacing), 2*radius+2*spacing))
	gc := draw2dimg.NewGraphicContext(img)

	for div := 0; div < divisions; div += 1 {
		startX := spacing + radius
		startY := spacing + radius

		rads := (totalRads / float64(divisions)) * float64(div)
		x := math.Cos(rads) * radius
		y := math.Sin(rads) * radius

		xInt := startX + int(x+0.5)
		yInt := startY + int(y+0.5)
		log.Printf("division %d = rads %f; --> %f,%f = (%d,%d) line (%d,%d)", div, rads, x, y, startX, startY, xInt, yInt)

		draw1PxLine(gc, image.Point{startX, startY}, image.Point{xInt, yInt}, color.Black)
		// gc.SetStrokeColor(color.Black)
		// gc.SetLineWidth(1.0)
		// gc.BeginPath()
		// gc.MoveTo(float64(startX), float64(startY))
		// gc.LineTo(float64(xInt), float64(yInt))
		// gc.Stroke()
	}

	return img
}

func lineBasic() *image.RGBA {
	const offset = 1
	const length = 10

	img := image.NewRGBA(image.Rect(0, 0, 2*length+2*offset, 2*length))
	gc := draw2dimg.NewGraphicContext(img)

	draw1PxLine(gc, image.Point{offset, length}, image.Point{length, length}, color.Black)
	draw1PxLine(gc, image.Point{2 * length, length}, image.Point{offset + length, length},
		color.RGBA{0x00, 0x00, 0xff, 0xff})
	return img
}
