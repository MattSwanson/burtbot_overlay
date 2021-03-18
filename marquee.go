package main

import (
	"image"
	"image/color"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/text"
)

type Marquee struct {
	on         bool
	speed      int
	x          int
	y          int
	text       string
	textBounds image.Rectangle
}

func (m *Marquee) enable(b bool) {
	m.on = b
}

func (m *Marquee) setText(s string) {
	m.text = s
	m.textBounds = text.BoundString(myFont, s)
	m.y = screenHeight - 2*m.textBounds.Dy() + 5
	m.on = true
}

func (m *Marquee) Update() error {
	m.x -= m.speed
	if m.x+m.textBounds.Dx() < 0 {
		m.x = screenWidth
	}
	return nil
}

func (m *Marquee) Draw(screen *ebiten.Image) {
	if m.on {
		text.Draw(screen, m.text, myFont, m.x+1, m.y+1, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
		text.Draw(screen, m.text, myFont, m.x, m.y, color.RGBA{0, 0xFF, 0, 0xFF})
	}

}
