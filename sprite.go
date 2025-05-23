package main

import (
	"math/rand"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

var sprites []rl.Texture2D

type Sprite struct {
	draw     bool
	posX     float64
	posY     float64
	scaleX   float64
	scaleY   float64
	width    float64 // Adjusted width based on scale and image width
	height   float64 // Adjusted height based on scale and image height
	vx       float64
	vy       float64
	objScale float64
	image    rl.Texture2D
}

type Sprites struct {
	sprites      []*Sprite
	num          int
	screenWidth  int
	screenHeight int
}

func LoadSprites() {
	sprites = []rl.Texture2D{}
	img := rl.LoadTexture("./images/BLUE_GOPHER.png")
	sprites = append(sprites, img)

	img = rl.LoadTexture("./images/green_goph.png")
	sprites = append(sprites, img)

	img = rl.LoadTexture("./images/tux_goph.png")
	sprites = append(sprites, img)
}

func NewSprite(sprite rl.Texture2D) Sprite {
	rvx := float64(rand.Intn(1280)) + 0.25
	rvy := float64(rand.Intn(720)) + 0.25
	return Sprite{
		draw:     true,
		posX:     0.0,
		posY:     0.0,
		scaleX:   1.0,
		scaleY:   1.0,
		objScale: 1.0,
		width:    float64(sprite.Width),
		height:   float64(sprite.Height),
		vx:       rvx,
		vy:       rvy,
		image:    sprite,
	}
}

func (o *Sprite) Update(delta float64) error {
	o.posX += o.vx * delta / 1000
	o.posY += o.vy * delta / 1000
	if o.posX <= 0 {
		o.vx = -o.vx
		o.posX = 0
	} else if o.posX >= screenWidth-o.width {
		o.vx = -o.vx
		o.posX = screenWidth - o.width
	} else if o.posY <= 0 {
		o.vy = -o.vy
		o.posY = 0
	} else if o.posY >= screenHeight-o.height {
		o.vy = -o.vy
		o.posY = screenHeight - o.height
	}
	return nil
}

func (o *Sprite) Draw() {
	if o.draw {
		rl.DrawTextureEx(o.image, rl.Vector2{X: float32(o.posX), Y: float32(o.posY)}, 0, float32(o.objScale), rl.White)
	}
}

// SetScale will set the object scale
func (o *Sprite) SetScale(scale float64) {
	o.objScale = scale
	o.width *= scale
	o.height *= scale
}

func (o *Sprite) SetPosition(x, y float64) {
	o.posX, o.posY = x, y
}
