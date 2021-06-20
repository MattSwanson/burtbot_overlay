package visuals

import (
	"time"

	"github.com/MattSwanson/burtbot_overlay/sound"
	rl "github.com/MattSwanson/raylib-go/raylib"
)

var img rl.Texture2D

const scaleSpeed = 2.0

type ErrorBox struct {
	Visible bool
	Scale   float32
}

func NewErrorBox() *ErrorBox {
	img = rl.LoadTexture("./images/hmm.png")
	return &ErrorBox{Scale: 3.0}
}

func (e *ErrorBox) ShowError() {
	// show the error we got
	e.Visible = true
	sound.Play("sosumi")
	go func() {
		time.Sleep(5 * time.Second)
		e.Visible = false
	}()
}

func (e *ErrorBox) Update(delta float64) {

}

func (e *ErrorBox) Draw() {
	if !e.Visible {
		return
	}
	rl.DrawTextureEx(img, rl.Vector2{X: 0, Y: 0}, 0.0, e.Scale, rl.White)
}
