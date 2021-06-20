package plinko

import (
	"fmt"
	"log"
	"math/rand"

	rl "github.com/MattSwanson/raylib-go/raylib"
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
	img         rl.Texture2D
	playerName  string
	playerColor rl.Color
	labelOffset fPoint
}

func NewToken(playerName, playerColor string, img rl.Texture2D, pos fPoint) *token {
	radius := float64(img.Width) / 2.0
	labelOffset := fPoint{2.0 * radius, 0}
	color, err := colorHexStrToColor(playerColor)
	if err != nil {
		log.Println("could not convert hex string to color", err.Error())
		color = rl.Blue
	}
	return &token{
		mass:        tokenMass,
		x:           pos.x,
		y:           pos.y,
		img:         img,
		radius:      radius,
		playerName:  playerName,
		playerColor: color,
		labelOffset: labelOffset,
	}
}

func colorHexStrToColor(colorString string) (rl.Color, error) {
	c := rl.Color{}
	c.A = 0xff
	_, err := fmt.Sscanf(colorString, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	return c, err
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

func (b *token) Draw() {
	if !b.falling {
		return
	}
	rl.DrawTexture(b.img, int32(b.x), int32(b.y), b.playerColor)
	rl.DrawText(b.playerName, int32(b.x+b.labelOffset.x), int32(b.y+b.labelOffset.y), 18, rl.Green)
}

func (b *token) Release() {
	r := rand.Float64()
	b.vx = (r - 0.5) * 3.0
	b.falling = true
}

func (b *token) SetVelocity(vx, vy float64) {
	b.vx, b.vy = vx, vy
}

func (b *token) SetPosition(x, y float64) {
	b.x, b.y = x, y
}
