package plinko

import (
	"fmt"
	"log"
	"math/big"
	"math/rand"

	"github.com/MattSwanson/burtbot_overlay/shaders"
	rl "github.com/MattSwanson/raylib-go/raylib"
)

const (
	tokenMass        = 5
	typeNormal       = 0x00
	typeSuper        = 0x01
	typeSecondChance = 0x02
)

type token struct {
	falling     bool
	mass        int
	tokenType   int
	x           float64
	y           float64
	vx          float64
	vy          float64
	radius      float64
	img         rl.Texture2D
	playerName  string
	playerColor rl.Color
	labelOffset fPoint
	shader      rl.Shader
	Value       *big.Int
}

// TODO: Update to specify a special token type to make and set the shader accordingly
func NewToken(playerName, playerColor string, img rl.Texture2D, pos fPoint, value *big.Int, tokenType int) *token {
	radius := float64(img.Width) / 2.0
	labelOffset := fPoint{2.0 * radius, 0}
	color, err := colorHexStrToColor(playerColor)
	if err != nil {
		log.Println("could not convert hex string to color", err.Error())
		color = rl.Blue
	}
	shader := rl.GetShaderDefault()
	switch tokenType {
	case typeSuper:
		shader = shaders.Get("cosmic")
		shaders.SetOffsets("cosmic", img.Width, img.Height)
	case typeSecondChance:
		shader = shaders.Get("secondChance")
		shaders.SetOffsets("secondChance", img.Width, img.Height)
	}

	//TODO: Set shader based on token created (super, second chance etc.)
	return &token{
		mass:        tokenMass,
		x:           pos.x,
		y:           pos.y,
		img:         img,
		tokenType:   tokenType,
		radius:      radius,
		playerName:  playerName,
		playerColor: color,
		labelOffset: labelOffset,
		shader:      shader,
		Value:       value,
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
	rl.BeginShaderMode(b.shader)
	rl.DrawTexture(b.img, int32(b.x), int32(b.y), b.playerColor)
	rl.EndShaderMode()
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
