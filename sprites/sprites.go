package sprites

import (
	"image"
	"image/color"
	"log"
	"math"

	"github.com/evanj/netgamesim/intersect"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
)

var tankGreen = color.RGBA{0x0c, 0xd4, 0x63, 0xff}
var targetDarkRed = color.RGBA{0x9a, 0x1f, 0x40, 0xff}
var targetLightRed = color.RGBA{0xd9, 0x45, 0x5f, 0xff}
var smokeGrey = color.RGBA{0xaa, 0xaa, 0xaa, 0xaa}

const TankSize = 20
const tankLineWidth = 2.0

const TargetSize = 30
const targetInnerSize = TargetSize / 2.0

const BulletSize = 12
const SmokeSize = 25

const pixelStrokeOffset = 0.5

func DrawTank(gc draw2d.GraphicContext, center intersect.Point) {
	gc.SetStrokeColor(color.Black)
	gc.SetFillColor(tankGreen)
	gc.SetLineWidth(tankLineWidth)
	// this should make the corners of the tank square but it has a bug:
	// https://github.com/llgcode/draw2d/issues/155
	// sadly this means we need to draw some insane lines with rectangles
	gc.SetLineJoin(draw2d.MiterJoin)

	// TODO: should this be for all odd line widths? E.g. 3.0, 5.0? what about 3.5 or 0.5?
	if tankLineWidth <= 1.0 {
		center.X += pixelStrokeOffset
		center.Y += pixelStrokeOffset
	}

	// tank "body"
	gc.MoveTo(center.X-TankSize/2, center.Y-TankSize/2)
	gc.LineTo(center.X+TankSize/2, center.Y-TankSize/2)
	gc.LineTo(center.X+TankSize/2, center.Y+TankSize/2)
	gc.LineTo(center.X-TankSize/2, center.Y+TankSize/2)
	gc.Close()

	// tank "gun"
	gc.MoveTo(center.X, center.Y)
	gc.LineTo(center.X+TankSize, center.Y)
	gc.FillStroke()

	gc.BeginPath()
}

func DrawTarget(gc draw2d.GraphicContext, center intersect.Point) {
	gc.SetFillColor(targetDarkRed)
	draw2dkit.Circle(gc, center.X, center.Y, TargetSize/2)
	gc.Fill()
	gc.SetFillColor(targetLightRed)
	draw2dkit.Circle(gc, center.X, center.Y, targetInnerSize/2)
	gc.Fill()
}

func DrawBullet(gc draw2d.GraphicContext, center intersect.Point) {
	gc.SetFillColor(color.Black)

	// the bullet is a line BulletSize wide and BulletSize/3 thick
	gc.SetLineWidth(BulletSize / 3)
	gc.BeginPath()
	gc.MoveTo(center.X-BulletSize/2, center.Y)
	gc.LineTo(center.X+BulletSize/2, center.Y)
	gc.Stroke()
	gc.BeginPath()
}

func DrawSmoke(gc draw2d.GraphicContext, center intersect.Point) {
	gc.SetFillColor(smokeGrey)
	draw2dkit.Circle(gc, center.X, center.Y, SmokeSize/2)
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

	for div := 0; div < divisions; div++ {
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

	for div := 0; div < divisions; div++ {
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
