package plinko

import (
	"image/color"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/text"
)

const (
	tokenMass = 5
)

type token struct {
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

func NewToken(playerName string, img *ebiten.Image, pos fPoint) *token {
	radius := float64(img.Bounds().Dx()) / 2.0
	labelOffset := fPoint{2.0 * radius, 0}
	return &token{
		mass:        tokenMass,
		x:           pos.x,
		y:           pos.y,
		img:         img,
		radius:      radius,
		playerName:  playerName,
		labelOffset: labelOffset,
	}
}

func (b *token) Update(delta float64) {
	if delta == 0 {
		return
	}
	if b.falling {
		b.vy = b.vy + gravity*delta/1000.0
		b.x += b.vx * delta / 1000.0
		b.y += b.vy * delta / 1000.0
	}
}

func (b *token) Draw(screen *ebiten.Image) {
	if !b.falling {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.x, b.y)
	screen.DrawImage(b.img, op)
	text.Draw(screen, b.playerName, playerLabelFont, int(b.x+b.labelOffset.x), int(b.y+b.labelOffset.y), color.RGBA{0x00, 0xff, 0x00, 0xff})
}

func (b *token) Release() {
	b.falling = true
}

func (b *token) SetVelocity(vx, vy float64) {
	b.vx, b.vy = vx, vy
}

func (b *token) SetPosition(x, y float64) {
	b.x, b.y = x, y
}
