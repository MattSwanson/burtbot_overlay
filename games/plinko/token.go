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
	Value       *big.Int
}

func NewToken(playerName, playerColor string, img rl.Texture2D, pos fPoint, value *big.Int) *token {
	radius := float64(img.Width) / 2.0
	labelOffset := fPoint{2.0 * radius, 0}
	color, err := colorHexStrToColor(playerColor)
	if err != nil {
		log.Println("could not convert hex string to color", err.Error())
		color = rl.Blue
	}
	// if value > 100 {
	// 	color = rl.White
	// }
	return &token{
		mass:        tokenMass,
		x:           pos.x,
		y:           pos.y,
		img:         img,
		radius:      radius,
		playerName:  playerName,
		playerColor: color,
		labelOffset: labelOffset,
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
	if b.Value.Cmp(big.NewInt(100)) == 1 {
		shaders.SetOffsets(b.img.Width, b.img.Height)
		rl.BeginShaderMode(shaders.Get("cosmic"))
		//loc := rl.GetShaderLocation(superShader, "u_time")
		//rl.SetShaderValue(superShader, loc, []float32{float32(time.Now().UnixNano())}, rl.ShaderUniformFloat)
	}
	rl.DrawTexture(b.img, int32(b.x), int32(b.y), b.playerColor)
	if b.Value.Cmp(big.NewInt(100)) == 1 {
		rl.EndShaderMode()
	}
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
