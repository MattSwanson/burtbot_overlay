package plinko

import "github.com/MattSwanson/ebiten/v2"

type ball struct {
	falling bool
	mass    int
	x       float64
	y       float64
	vx      float64
	vy      float64
	radius  float64
	img     *ebiten.Image
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
}

func (b *ball) Release() {
	b.falling = true
}

func (b *ball) SetVelocity(vx, vy float64) {
	b.vx, b.vy = vx, vy
}
