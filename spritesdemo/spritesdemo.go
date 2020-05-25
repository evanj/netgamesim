package main

import (
	"image"
	"image/png"
	"os"

	"github.com/evanj/netgamesim/intersect"
	"github.com/evanj/netgamesim/sprites"
	"github.com/llgcode/draw2d/draw2dimg"
)

func writePNG(path string, img image.Image) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func main() {
	img := image.NewRGBA(image.Rect(0, 0, int(sprites.TankSize*2), int(sprites.TankSize*2)))
	gc := draw2dimg.NewGraphicContext(img)
	sprites.DrawTank(gc, intersect.Point{X: sprites.TankSize, Y: sprites.TankSize})
	err := writePNG("tank.png", img)
	if err != nil {
		panic(err)
	}
	// err = writePNG("angles.png", s.Angles)
	// if err != nil {
	// 	panic(err)
	// }
	// err = writePNG("angles2.png", s.Angles2)
	// if err != nil {
	// 	panic(err)
	// }
	// err = writePNG("lines.png", s.Lines)
	// if err != nil {
	// 	panic(err)
	// }
}
