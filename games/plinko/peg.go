package plinko

import "github.com/MattSwanson/ebiten/v2"

type peg struct {
	drain  int
	x      float64
	y      float64
	radius float64
	img    *ebiten.Image
}

// func (p *peg) Update(delta float64) {
// 	p.x += p.vx * delta / 1000.0
// }

func (p *peg) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.x, p.y)
	screen.DrawImage(p.img, op)
}
