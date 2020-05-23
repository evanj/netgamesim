package main

import (
	"image"
	"image/png"
	"os"

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
	s := sprites.New()

	img := image.NewRGBA(image.Rect(0, 0, int(s.Tank.Size*2), int(s.Tank.Size*2)))
	gc := draw2dimg.NewGraphicContext(img)
	s.Tank.Draw(gc, s.Tank.Size, s.Tank.Size)
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
