package main

import (
	"errors"
	"image"
	"image/color"
	"log"
	"math/rand"
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
	speed       float64
	x           float64
	y           int
	yOffset     int
	text        string
	textBounds  image.Rectangle
	color       color.RGBA
	currentFont *font.Face
	oneShot     bool
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

func NewMarquee(speed float64, color color.RGBA, oneShot bool) *Marquee {
	var currentFont *font.Face
	if rand.Intn(100) < 10 {
		currentFont = &marqueeFontXl
	} else {
		currentFont = &marqueeFont
	}
	return &Marquee{speed: speed, currentFont: currentFont, color: color, oneShot: oneShot}
}

func (m *Marquee) enable(b bool) {
	m.on = b
}

func (m *Marquee) setText(s string) {
	m.text = s
	m.textBounds = text.BoundString(*m.currentFont, s)
	// 0 to screenHeight - m.textBounds.Dy() + m.yOffset
	m.y = rand.Intn(screenHeight-m.textBounds.Dy()) + m.textBounds.Dy()
	m.color = color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 0xff}
	m.x = screenWidth
	m.on = true
}

func (m *Marquee) Update(delta float64) error {
	m.x -= m.speed * delta / 1000.0
	if m.x+float64(m.textBounds.Dx()) < 0 {
		if m.oneShot {
			return errors.New("i'm done")
		} else {
			m.x = screenWidth
		}
	}
	return nil
}

func (m *Marquee) Draw(screen *ebiten.Image) {
	if m.on {
		//text.Draw(screen, m.text, *m.currentFont, m.x+1, m.y+1, m.color)
		text.Draw(screen, m.text, *m.currentFont, int(m.x), m.y, m.color)
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

func (m *Marquee) SetSpeed(speed float64) {
	m.speed = speed
}
