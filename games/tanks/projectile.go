package tanks

import (
	rl "github.com/MattSwanson/raylib-go/raylib"
)

const gravity float64 = 500.0
const radius float64 = 8.0

type projectile struct {
	x      float64
	y      float64
	vx     float64
	vy     float64
	radius float64
	img    rl.Texture2D
	wv     float64
	marker bool
}

func NewProjectile(x, y float64, wind float64, marker bool) *projectile {
	img := rl.GenImageColor(int(radius*2), int(radius*2), rl.Color{R: 255, G: 0, B: 0, A: 255})
	return &projectile{
		x:      x,
		y:      y,
		img:    rl.LoadTextureFromImage(img),
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
	p.x += p.vx * delta / 1000.0
	p.y += p.vy * delta / 1000.0
}

func (p *projectile) Draw() {
	rl.DrawTexture(p.img, int32(p.x), int32(p.y), rl.Red)
}

func (p *projectile) SetVelocity(x, y float64) {
	p.vx, p.vy = x, y
}
