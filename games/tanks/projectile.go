package tanks

import (
	rl "github.com/MattSwanson/raylib-go/raylib"
)

const gravity float64 = 500.0
const radius float64 = 8.0
const trailLength = 50

type projectile struct {
	x      float64
	y      float64
	prevXs []float64
	prevYs []float64
	vx     float64
	vy     float64
	radius float64
	wv     float64
	marker bool
}

func NewProjectile(x, y float64, wind float64, marker bool) *projectile {
	prevXs := make([]float64, trailLength)
	prevYs := make([]float64, trailLength)
	for i := 0; i < trailLength; i++ {
		prevXs = append(prevXs, x)
		prevYs = append(prevYs, y)
	}
	return &projectile{
		x:      x,
		y:      y,
		prevXs: prevXs,
		prevYs: prevYs,
		wv:     wind,
		radius: radius,
		marker: marker,
	}
}

func (p *projectile) Update(delta float64) {
	if p.marker {
		return
	}
	p.vx = p.vx + p.wv*delta/1000.0
	p.vy = p.vy + gravity*delta/1000.0

	p.prevXs = append(p.prevXs[1:], p.x)
	p.prevYs = append(p.prevYs[1:], p.y)

	p.x += p.vx * delta / 1000.0
	p.y += p.vy * delta / 1000.0

}

func (p *projectile) Draw() {
	for i := 0; i < len(p.prevXs); i++ {
		a := uint8(float32(i)/float32(len(p.prevXs)) * 255)
		//a := uint8((1.0 - (1.0/float32(trailLength))*float32(len(p.prevXs)-i)) * 255)
		rl.DrawCircle(int32(p.prevXs[i]), int32(p.prevYs[i]), float32(p.radius * float64(i)/float64(len(p.prevXs))), rl.Color{R: 0xff, G: 0x00, B: 0x00, A: a})
	}
	rl.DrawCircle(int32(p.x), int32(p.y), float32(p.radius), rl.Color{R: 0xff, G: 0x00, B: 0x00, A: 0xff})
}

func (p *projectile) SetVelocity(x, y float64) {
	p.vx, p.vy = x, y
}
