package plinko

import (
	rl "github.com/MattSwanson/raylib-go/raylib"
)

type peg struct {
	x      float64
	y      float64
	radius float64
	img    rl.Texture2D
}

func (p *peg) Draw() {
	rl.DrawTexture(p.img, int32(p.x), int32(p.y), rl.White)
}
