package plinko

import (
	rl "github.com/MattSwanson/raylib-go/raylib"
)

type barrier struct {
	x       float64
	y       float64
	w       float64
	h       float64
	rebound float64
	sprite  rl.Texture2D
	bounds  []edge
}

func NewBarrier(sprite rl.Texture2D) *barrier {
	b := barrier{
		sprite: sprite,
		w:      float64(sprite.Width),
		h:      float64(sprite.Height),
		bounds: []edge{},
	}
	return &b
}

func (b *barrier) SetPosition(x, y float64) {
	b.x, b.y = x, y
	b.generateBounds()
}

func (b *barrier) generateBounds() {
	e2 := edge{b.x - b.w/2, b.y + b.h/2, b.x + b.w/2, b.y + b.h/2}
	e1 := edge{b.x, b.y - b.h/2, b.x - b.w/2, b.y + b.h/2}
	e0 := edge{b.x + b.w/2, b.y + b.h/2, b.x, b.y - b.h/2}
	b.bounds = []edge{e0, e1, e2}
}

func (b *barrier) Draw() {
	rl.DrawTexture(b.sprite, int32(b.x-b.w/2), int32(b.y-b.h/2), rl.White)
}

func (b *barrier) GetRebound() float64 {
	return b.rebound
}

// edge should be defined in ccw manner
type edge struct {
	x0 float64
	y0 float64
	x1 float64
	y1 float64
}

func (e edge) getMidpoint() (float64, float64) {
	return (e.x0 + e.x1) / 2, (e.y0 + e.y1) / 2
}

func (e edge) IsLeft(x, y float64) float64 {
	return (e.x1-e.x0)*(y-e.y0) -
		(x-e.x0)*(e.y1-e.y0)
}
