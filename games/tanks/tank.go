package tanks

import (
	"image/color"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/text"
)

type tank struct {
	playerName string
	x          float64
	y          float64
	img        *ebiten.Image
}

func NewTank(playerName string) *tank {
	img := ebiten.NewImage(64, 64)
	img.Fill(color.RGBA{0x00, 0x00, 0xff, 0xff})
	return &tank{
		playerName: playerName,
		img:        img,
	}
}

func (t *tank) setPosition(x, y float64) {
	t.x, t.y = x, y
}

func (t *tank) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(t.x, t.y)
	screen.DrawImage(t.img, op)
	text.Draw(screen, t.playerName, playerLabelFont, int(t.x), int(t.y), color.RGBA{0x00, 0xff, 0x00, 0xff})
}
