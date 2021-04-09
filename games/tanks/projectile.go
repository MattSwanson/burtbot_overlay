package tanks

import (
	"image/color"

	"github.com/MattSwanson/ebiten/v2"
)

const gravity float64 = 500.0
const radius float64 = 8.0

type projectile struct {
	x      float64
	y      float64
	vx     float64
	vy     float64
	radius float64
	img    *ebiten.Image
	wv     float64
}

func NewProjectile(x, y float64, wind float64) *projectile {
	img := ebiten.NewImage(int(radius*2), int(radius*2))
	img.Fill(color.RGBA{0xff, 0x00, 0x00, 0xff})
	return &projectile{
		x:      x,
		y:      y,
		img:    img,
		wv:     wind,
		radius: radius,
	}
}

func (p *projectile) Update(delta float64) {
	p.vx = p.vx + p.wv*delta/1000.0
	p.vy = p.vy + gravity*delta/1000.0
	p.x += p.vx * delta / 1000.0
	p.y += p.vy * delta / 1000.0
}

func (p *projectile) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.x-radius, p.y-radius)
	screen.DrawImage(p.img, &op)
}

func (p *projectile) SetVelocity(x, y float64) {
	p.vx, p.vy = x, y
}
