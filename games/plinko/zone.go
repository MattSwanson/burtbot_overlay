package plinko

import (
	"fmt"
	"image/color"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/text"
)

type zone struct {
	x           float64
	y           float64
	w           float64
	h           float64
	rewardValue int
	img         *ebiten.Image
}

func NewZone(rect fRect, n int) *zone {
	img := ebiten.NewImage(int(rect.Dx()), int(rect.Dy()))
	r := uint8(float64(n) / 10.0 * 255.0)
	img.Fill(color.RGBA{r, 0x00, 0x00, 0x33})
	return &zone{
		x:           rect.min.x,
		y:           rect.min.y,
		w:           rect.Dx(),
		h:           rect.Dy(),
		rewardValue: n,
		img:         img,
	}
}

func (z *zone) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(z.x, z.y)
	screen.DrawImage(z.img, op)
	text.Draw(screen, fmt.Sprint(z.rewardValue), gameFont, int(z.x+z.w/2), int(z.y+z.h/2), color.RGBA{0x00, 0xff, 0x00, 0x55})
}
