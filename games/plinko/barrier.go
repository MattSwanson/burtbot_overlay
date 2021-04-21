package plinko

import (
	"image/color"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/ebitenutil"
)

type barrier struct {
	x       float64
	y       float64
	w       float64
	h       float64
	rebound float64
	sprite  *ebiten.Image
	collImg *ebiten.Image
	bounds  []edge
}

func NewBarrier(sprite *ebiten.Image) *barrier {
	b := barrier{
		sprite: sprite,
		w:      float64(sprite.Bounds().Dx()),
		h:      float64(sprite.Bounds().Dy()),
		bounds: []edge{},
	}
	return &b
}

func (b *barrier) SetPosition(x, y float64) {
	b.x, b.y = x, y
	b.generateBounds()
}

func (b *barrier) generateBounds() {
	e2 := edge{b.x - b.w/2, b.y + b.h/2, b.x + b.w/2, b.y + b.h/2}
	e1 := edge{b.x, b.y - b.h/2, b.x - b.w/2, b.y + b.h/2}
	e0 := edge{b.x + b.w/2, b.y + b.h/2, b.x, b.y - b.h/2}
	b.bounds = []edge{e0, e1, e2}
	b.collImg = drawCollisionArea(b)
}

func (b *barrier) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.x-b.w/2, b.y-b.h/2)
	screen.DrawImage(b.sprite, &op)
	// for _, e := range b.bounds {
	// 	e.DrawNormal(screen)
	// }
	//ebitenutil.DrawLine(screen, b.bounds[0].x0, b.bounds[0].y0, b.bounds[0].x1, b.bounds[0].y1, color.RGBA{0xff, 0x00, 0x00, 0xff})
	//ebitenutil.DrawLine(screen, b.bounds[1].x0, b.bounds[1].y0, b.bounds[1].x1, b.bounds[1].y1, color.RGBA{0xff, 0x00, 0x00, 0xff})
	screen.DrawImage(b.collImg, nil)
}

func (b *barrier) GetRebound() float64 {
	return b.rebound
}

type bounds []edge

// edge should be defined in ccw manner
type edge struct {
	x0 float64
	y0 float64
	x1 float64
	y1 float64
}

func (e edge) getMidpoint() (float64, float64) {
	return (e.x0 + e.x1) / 2, (e.y0 + e.y1) / 2
}

// P0 =
func (e edge) IsLeft(x, y float64) float64 {
	return (e.x1-e.x0)*(y-e.y0) -
		(x-e.x0)*(e.y1-e.y0)
}

func (e edge) DrawNormal(screen *ebiten.Image) {
	mpx, mpy := e.getMidpoint()
	dx := e.x1 - e.x0
	dy := e.y1 - e.y0
	// if dx < 0 and dy < 0 then flip x and keep y
	sx, sy := 1.0, 1.0
	if dx < 0 && dy < 0 {
		sx = -sx
	} else if dx < 0 && dy > 0 {
		sy = -sy
	}
	// if dx < 0 and dy > 0 then flip both
	// if dx
	nx := sx * (e.x1 - e.x0)
	ny := sy * (e.y1 - e.y0)
	ebitenutil.DrawLine(screen, mpx, mpy, mpx+nx, mpy+ny, color.RGBA{0x00, 0xFF, 0x00, 0xFF})
}

func drawCollisionArea(b *barrier) *ebiten.Image {
	collImg := ebiten.NewImage(2560, 1440)
	bytes := make([]byte, 2560*1440*4)
	//maxDist := math.Sqrt(t.w*t.w+t.h*t.h) + 8
	for x := 0; x < 2560; x++ {
		for y := 0; y < 1440; y++ {
			if b.bounds[0].IsLeft(float64(x), float64(y)) < 0 && b.bounds[1].IsLeft(float64(x), float64(y)) < 0 && b.bounds[2].IsLeft(float64(x), float64(y)) < 0 {
				// set green
				bytes[(2560*y*4)+(x*4)+3] = 0xff
				bytes[(2560*y*4)+(x*4)+1] = 0xff
			}
		}
	}
	collImg.ReplacePixels(bytes)
	return collImg
}
