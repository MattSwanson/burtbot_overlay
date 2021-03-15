package main

import (
	"image"

	"github.com/MattSwanson/ebiten/v2"
)

type rando struct {
	x, y, z, w uint32
}

type static struct {
	noiseImage *image.RGBA
}

func (r *rando) next() uint32 {
	t := r.x ^ (r.x << 11)
	r.x, r.y, r.z = r.y, r.z, r.w
	r.w = (r.w ^ (r.w >> 19)) ^ (t ^ (t >> 8))
	return r.w
}

var theRand = &rando{12345678, 4185243, 776511, 45411}

func (s *static) Update() error {
	const l = screenWidth * screenHeight
	for i := 0; i < l; i++ {
		x := theRand.next()
		s.noiseImage.Pix[4*i] = uint8(x >> 24)
		s.noiseImage.Pix[4*i+1] = uint8(x >> 16)
		s.noiseImage.Pix[4*i+2] = uint8(x >> 8)
		s.noiseImage.Pix[4*i+3] = 0xff
	}
	return nil
}

func (s *static) Draw(screen *ebiten.Image) {
	screen.ReplacePixels(s.noiseImage.Pix)
}
