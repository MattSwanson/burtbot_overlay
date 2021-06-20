package plinko

import (
	"fmt"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

type zone struct {
	x           float64
	y           float64
	w           float64
	h           float64
	rewardValue int
	hits        int
	img         rl.Texture2D
}

func NewZone(rect fRect, n int) *zone {
	r := uint8(float64(n) / 10.0 * 255.0)
	img := rl.GenImageColor(int(rect.Dx()), int(rect.Dy()), rl.Color{r, 0x00, 0x00, 0x33})
	return &zone{
		x:           rect.min.x,
		y:           rect.min.y,
		w:           rect.Dx(),
		h:           rect.Dy(),
		rewardValue: n,
		img:         rl.LoadTextureFromImage(img),
	}
}

func (z *zone) Draw() {
	rl.DrawTexture(z.img, int32(z.x), int32(z.y), rl.White)
	rl.DrawText(fmt.Sprint(z.rewardValue), int32(z.x+z.w/2), int32(gameHeight-80), 64, rl.Color{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF})
}

func (z *zone) Hit() {
	z.hits++
}
