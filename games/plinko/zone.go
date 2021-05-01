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
	// op := &ebiten.DrawImageOptions{}
	// op.GeoM.Translate(z.x, z.y)
	// screen.DrawImage(z.img, op)
	// text.Draw(screen, fmt.Sprint(z.rewardValue), gameFont, int(z.x+z.w/2), int(gameHeight), color.RGBA{0x00, 0xff, 0x00, 0x55})
	// text.Draw(screen, fmt.Sprint(z.hits), gameFont, int(z.x+z.w/2), int(z.y+3*z.h/2), color.RGBA{0x00, 0xff, 0x00, 0x11})
	rl.DrawTexture(z.img, int32(z.x), int32(z.y), rl.White)
	rl.DrawText(fmt.Sprint(z.rewardValue), int32(z.x+z.w/2), int32(gameHeight-80), 64, rl.Color{R: 0x00, G: 0xFF, B: 0x00, A: 0x55})
}

func (z *zone) Hit() {
	z.hits++
}
