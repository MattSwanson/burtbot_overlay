package main

import (
	"image"
	"image/color"
	"log"
	"os"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var marqueeFont font.Face
var marqueeFontXl font.Face

const (
	xlYOffset  = screenHeight / 2
	regYOffset = -5
)

type Marquee struct {
	on          bool
	speed       int
	x           int
	y           int
	yOffset     int
	text        string
	textBounds  image.Rectangle
	currentFont *font.Face
}

func init() {
	// font init
	bs, err := os.ReadFile("caskaydia.TTF")
	if err != nil {
		log.Fatal(err)
	}
	tt, err := opentype.Parse(bs)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	marqueeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    96,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	const xldpi = 144
	marqueeFontXl, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    512,
		DPI:     xldpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func NewMarquee(speed int) *Marquee {
	return &Marquee{speed: speed, currentFont: &marqueeFont}
}

func (m *Marquee) enable(b bool) {
	m.on = b
}

func (m *Marquee) setText(s string) {
	m.text = s
	m.textBounds = text.BoundString(*m.currentFont, s)
	m.y = screenHeight - m.textBounds.Dy() + m.yOffset
	m.x = screenWidth
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
		text.Draw(screen, m.text, *m.currentFont, m.x+1, m.y+1, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
		text.Draw(screen, m.text, *m.currentFont, m.x, m.y, color.RGBA{0, 0xFF, 0, 0xFF})
	}
}

func (m *Marquee) Embiggen() {
	m.on = false
	m.currentFont = &marqueeFontXl
	m.yOffset = xlYOffset
	m.setText(m.text)
}

func (m *Marquee) Smol() {
	m.on = false
	m.currentFont = &marqueeFont
	m.yOffset = regYOffset
	m.setText(m.text)
}

func (m *Marquee) SetSpeed(speed int) {
	m.speed = speed
}
