package plinko

import "github.com/MattSwanson/ebiten/v2"

type peg struct {
	x      float64
	y      float64
	radius float64
	img    *ebiten.Image
}

func (p *peg) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.x, p.y)
	screen.DrawImage(p.img, op)
}
