package plinko

import (
	"image/color"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/text"
)

const (
	ballMass = 5
)

type ball struct {
	falling     bool
	mass        int
	x           float64
	y           float64
	vx          float64
	vy          float64
	radius      float64
	img         *ebiten.Image
	playerName  string
	labelOffset fPoint
}

func NewBall(playerName string, img *ebiten.Image, pos fPoint) *ball {
	radius := float64(img.Bounds().Dx()) / 2.0
	labelOffset := fPoint{2.0 * radius, 0}
	return &ball{
		mass:        ballMass,
		x:           pos.x,
		y:           pos.y,
		img:         img,
		radius:      radius,
		playerName:  playerName,
		labelOffset: labelOffset,
	}
}

func (b *ball) Update(delta float64) {
	if delta == 0 {
		return
	}
	if b.falling {
		b.vy = b.vy + gravity*delta/1000.0
		b.x += b.vx * delta / 1000.0
		b.y += b.vy * delta / 1000.0
	}
}

func (b *ball) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.x, b.y)
	screen.DrawImage(b.img, op)
	text.Draw(screen, b.playerName, playerLabelFont, int(b.x+b.labelOffset.x), int(b.y+b.labelOffset.y), color.RGBA{0x00, 0xff, 0x00, 0xff})
}

func (b *ball) Release() {
	b.falling = true
}

func (b *ball) SetVelocity(vx, vy float64) {
	b.vx, b.vy = vx, vy
}

func (b *ball) SetPosition(x, y float64) {
	b.x, b.y = x, y
}
