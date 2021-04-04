package tanks

import (
	"image/color"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/text"
)

type tank struct {
	playerName               string
	x                        float64
	y                        float64
	w                        float64
	h                        float64
	a                        float64
	projectileOffsetDistance float64
	img                      *ebiten.Image
}

func NewTank(playerName string) *tank {
	img := ebiten.NewImage(48, 48)
	img.Fill(color.RGBA{0x00, 0x00, 0xff, 0xff})
	return &tank{
		playerName:               playerName,
		img:                      img,
		w:                        float64(img.Bounds().Dx()),
		h:                        float64(img.Bounds().Dy()),
		projectileOffsetDistance: 60,
	}
}

func (t *tank) setPosition(x, y float64) {
	t.x, t.y = x, y
}

func (t *tank) setAngle(theta float64) {
	t.a = theta
}

func (t *tank) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-t.w/2, -t.h)
	op.GeoM.Rotate(t.a)
	op.GeoM.Translate(t.w/2, t.h)
	op.GeoM.Translate(t.x, t.y)
	screen.DrawImage(t.img, op)
	text.Draw(screen, t.playerName, playerLabelFont, int(t.x), int(t.y), color.RGBA{0xff, 0x00, 0x00, 0xff})
}
