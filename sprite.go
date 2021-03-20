package main

import (
	"log"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/ebitenutil"
)

var sprites []*ebiten.Image

type Sprite struct {
	draw     bool
	posX     float64
	posY     float64
	scaleX   float64
	scaleY   float64
	objScale float64
	image    *ebiten.Image
}

type Sprites struct {
	sprites []*Sprite
	num     int
}

func init() {
	// load images
	sprites = []*ebiten.Image{}
	img, _, err := ebitenutil.NewImageFromFile("./images/BLUE_GOPHER.png")
	if err != nil {
		log.Fatal(err)
	}
	sprites = append(sprites, img)

	img, _, err = ebitenutil.NewImageFromFile("./images/green_goph.png")
	if err != nil {
		log.Fatal(err)
	}
	sprites = append(sprites, img)

	img, _, err = ebitenutil.NewImageFromFile("./images/tux_goph.png")
	if err != nil {
		log.Fatal(err)
	}
	sprites = append(sprites, img)
}

func NewSprite(sprite *ebiten.Image) Sprite {
	return Sprite{true, 0.0, 0.0, 1.0, 1.0, 1.0, sprite}
}

func (o *Sprite) Update() error {
	return nil
}

func (o *Sprite) Draw(screen *ebiten.Image) {
	if o.draw {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(o.posX/o.objScale, o.posY/o.objScale)
		op.GeoM.Scale(o.objScale, o.objScale)
		screen.DrawImage(o.image, op)
	}
}

func (o *Sprite) SetScale(scale float64) {
	o.objScale = scale
}

func (o *Sprite) SetPosition(x, y float64) {
	o.posX, o.posY = x, y
}
