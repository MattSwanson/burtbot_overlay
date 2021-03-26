package plinko

import (
	"image/color"

	"github.com/MattSwanson/ebiten/v2"
)

type box struct {
	x        float64
	y        float64
	w        float64
	h        float64
	friction float64
	img      *ebiten.Image
}

func NewBox(bounds fRect, color color.RGBA) *box {
	img := ebiten.NewImage(int(bounds.Dx()), int(bounds.Dy()))
	img.Fill(color)
	return &box{
		img: img,
		x:   float64(bounds.min.x),
		y:   float64(bounds.min.y),
		w:   float64(bounds.Dx()),
		h:   float64(bounds.Dy()),
	}
}

func (b *box) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.x, b.y)
	screen.DrawImage(b.img, op)
}
