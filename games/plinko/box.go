package plinko

import (
	"image/color"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

type box struct {
	x   float64
	y   float64
	w   float64
	h   float64
	img rl.Texture2D
}

func NewBox(bounds fRect, color color.RGBA) *box {
	img := rl.GenImageColor(int(bounds.Dx()), int(bounds.Dy()), rl.Color{R: color.R, G: color.G, B: color.B, A: color.A})
	return &box{
		img: rl.LoadTextureFromImage(img),
		x:   float64(bounds.min.x),
		y:   float64(bounds.min.y),
		w:   float64(bounds.Dx()),
		h:   float64(bounds.Dy()),
	}
}

func (b *box) Draw() {
	rl.DrawTexture(b.img, int32(b.x), int32(b.y), rl.White)
}
