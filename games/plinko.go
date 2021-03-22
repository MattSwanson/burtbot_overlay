package games

import (
	"image/color"
	"log"
	"math"
	"time"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/audio"
	"github.com/MattSwanson/ebiten/v2/ebitenutil"
)

// win tokens by spedning tokens
// collisions? resolution on circular collisions

const (
	gravity    float64 = 500.0
	gameHeight float64 = 1300
	gameWidth  float64 = 2560
	numRows    int     = 13
	numColumns int     = 19
)

var initialBallPositionX float64

type Plinko struct {
	lastUpdate       time.Time
	ball             *ball
	pegs             []*peg
	boxes            []*box
	sounds           map[string]*audio.Player
	dropPoints       []fPoint
	currentDropPoint int
}

type fPoint struct {
	x float64
	y float64
}

type fRect struct {
	min fPoint
	max fPoint
}

func (r *fRect) Dx() float64 {
	return r.max.x - r.min.x
}

func (r *fRect) Dy() float64 {
	return r.max.y - r.min.y
}

func LoadPlinko(screenWidth, screenHeight float64, sounds map[string]*audio.Player) *Plinko {
	ballImg, _, err := ebitenutil.NewImageFromFile("./images/plinko/blue_token.png")
	if err != nil {
		log.Fatal(err)
	}
	pegImg, _, err := ebitenutil.NewImageFromFile("./images/plinko/ball.png")
	if err != nil {
		log.Fatal(err)
	}
	pegs := make([]*peg, numColumns*numRows)
	halfImgWidth := float64(ballImg.Bounds().Dx()) / 2.0
	offset := 25.0
	for i := 0; i < len(pegs); i++ {
		if i%numColumns == 0 {
			offset *= -1
		}
		p := peg{
			x:      (float64((i%numColumns)-numColumns/2)*100.0 + (screenWidth/2 - halfImgWidth)) + offset,
			y:      float64(i/numColumns)*75.0 + 200.0,
			radius: float64(pegImg.Bounds().Dx()) / 2.0,
			img:    pegImg,
		}
		pegs[i] = &p
	}
	dropPoints := []fPoint{}
	for i := 0; i < 5; i++ {
		dropPoints = append(dropPoints, fPoint{(screenWidth/2 - halfImgWidth) + (float64(i)-2)*300, 20.0})
	}
	//initialBallPositionX = screenWidth/2 - halfImgWidth
	b := ball{
		mass:   5,
		x:      dropPoints[2].x,
		y:      dropPoints[2].y,
		img:    ballImg,
		radius: float64(ballImg.Bounds().Dx()) / 2.0,
	}
	boxes := make([]*box, 3)
	boxHeight := 0.05 * gameHeight
	boxWidth := gameWidth
	x := 0.5 * (screenWidth - gameWidth)
	y := gameHeight - boxHeight + 0.5*(screenHeight-gameHeight)
	bounds := fRect{fPoint{x, y}, fPoint{x + boxWidth, y + boxHeight}}
	boxone := NewBox(bounds, color.RGBA{0x00, 0x00, 0xff, 0xff})
	boxes[0] = boxone

	boxHeight = gameHeight
	boxWidth = 0.025 * gameWidth
	x = 0.5 * (screenWidth - gameWidth)
	y = gameHeight - boxHeight + 0.5*(screenHeight-gameHeight)
	bounds = fRect{fPoint{x, y}, fPoint{x + boxWidth, y + boxHeight}}
	boxtwo := NewBox(bounds, color.RGBA{0x00, 0x00, 0xff, 0xff})
	boxes[1] = boxtwo

	boxHeight = gameHeight
	boxWidth = 0.025 * gameWidth

	x = screenWidth - boxWidth - 0.5*(screenWidth-gameWidth)
	y = gameHeight - boxHeight + 0.5*(screenHeight-gameHeight)
	bounds = fRect{fPoint{x, y}, fPoint{x + boxWidth, y + boxHeight}}
	boxthree := NewBox(bounds, color.RGBA{0x00, 0x00, 0xff, 0xff})
	boxes[2] = boxthree
	return &Plinko{ball: &b, pegs: pegs, boxes: boxes, sounds: sounds, currentDropPoint: 2, dropPoints: dropPoints}
}

func (p *Plinko) MoveDropPoint(direction int) {
	if direction == -1 && p.currentDropPoint > 0 {
		p.currentDropPoint--
		p.ball.x, p.ball.y = p.dropPoints[p.currentDropPoint].x, p.dropPoints[p.currentDropPoint].y
	} else if direction == 1 && p.currentDropPoint < len(p.dropPoints)-1 {
		p.currentDropPoint++
		p.ball.x, p.ball.y = p.dropPoints[p.currentDropPoint].x, p.dropPoints[p.currentDropPoint].y
	}
}

type ball struct {
	falling bool
	mass    int
	x       float64
	y       float64
	vx      float64
	vy      float64
	radius  float64
	img     *ebiten.Image
}

func (b *ball) Update(delta float64) {
	if delta == 0 {
		return
	}
	if b.falling {
		b.vy = b.vy + gravity*delta/1000.0
		b.x += b.vx * delta / 1000.0
		b.y += b.vy * delta / 1000.0
	}
}

func (p *Plinko) CheckForCollision() {
	const drain float64 = 0.95
	for _, peg := range p.pegs {
		dx := p.ball.x - peg.x
		dy := p.ball.y - peg.y
		mag := math.Sqrt(dx*dx + dy*dy)
		vmag := math.Sqrt(p.ball.vx*p.ball.vx + p.ball.vy*p.ball.vy)
		if mag <= peg.radius+p.ball.radius {
			p.ball.vx = (drain * vmag) * (dx / mag)
			p.ball.vy = (drain * vmag) * (dy / mag)
			// .05 -> .25
			//p.sounds["bip"].SetVolume()
			p.sounds["bip"].Rewind()
			p.sounds["bip"].Play()
		}
	}
	for _, box := range p.boxes {
		dY := math.Abs((box.y + 0.5*box.h) - (p.ball.y + p.ball.radius))
		dX := math.Abs((box.x + 0.5*box.w) - (p.ball.x + p.ball.radius))
		if dY < 0.5*box.h+p.ball.radius && dX < box.w/2 {
			p.ball.vy = -p.ball.vy * 0.6
			p.ball.vx = p.ball.vx * 0.6
			p.ball.y = box.y - 2.0*p.ball.radius
			if math.Abs(p.ball.vy) > 25 {
				p.sounds["boing"].Rewind()
				p.sounds["boing"].Play()
			}
		} else if dX < p.ball.radius+0.5*box.w && dY < box.h/2-p.ball.radius {
			p.ball.vx = -p.ball.vx * 0.6
			if p.ball.vx > 0 {
				p.ball.x = box.x + box.w
			} else {
				p.ball.x = box.x - 2.0*p.ball.radius
			}
			p.ball.vy = p.ball.vy * 0.6
			if math.Abs(p.ball.vx) > 25 {
				p.sounds["boing"].Rewind()
				p.sounds["boing"].Play()
			}
		}
	}
}

func (b *ball) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.x, b.y)
	screen.DrawImage(b.img, op)
}

func (b *ball) Release() {
	b.falling = true
}

func (b *ball) SetVelocity(vx, vy float64) {
	b.vx, b.vy = vx, vy
}

type peg struct {
	drain  int
	x      float64
	y      float64
	radius float64
	img    *ebiten.Image
}

// func (p *peg) Update(delta float64) {
// 	p.x += p.vx * delta / 1000.0
// }

func (p *peg) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.x, p.y)
	screen.DrawImage(p.img, op)
}

type box struct {
	x        float64
	y        float64
	w        float64
	h        float64
	friction float64
	img      *ebiten.Image
}

func NewBox(bounds fRect, color color.RGBA) *box {
	img := ebiten.NewImage(int(bounds.Dx()), int(bounds.Dy()))
	img.Fill(color)
	return &box{
		img: img,
		x:   float64(bounds.min.x),
		y:   float64(bounds.min.y),
		w:   float64(bounds.Dx()),
		h:   float64(bounds.Dy()),
	}
}

func (b *box) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.x, b.y)
	screen.DrawImage(b.img, op)
}

func (p *Plinko) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		p.ResetGame()
		return nil
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) && !p.ball.falling {
		p.MoveDropPoint(p.currentDropPoint - 1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) && !p.ball.falling {
		p.MoveDropPoint(p.currentDropPoint + 1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) && !p.ball.falling {
		p.ReleaseBall()
	}
	p.ball.Update(float64(time.Since(p.lastUpdate).Milliseconds()))
	p.CheckForCollision()
	p.lastUpdate = time.Now()
	return nil
}

func (p *Plinko) Draw(screen *ebiten.Image) {
	p.ball.Draw(screen)
	for _, v := range p.pegs {
		v.Draw(screen)
	}
	for _, v := range p.boxes {
		v.Draw(screen)
	}

}

func (p *Plinko) ResetGame() {
	p.ball.falling = false
	p.ball.vx = 0
	p.ball.vy = 0
	p.currentDropPoint = 2
	p.ball.x = p.dropPoints[p.currentDropPoint].x
	p.ball.y = p.dropPoints[p.currentDropPoint].y
}

func (p *Plinko) ReleaseBall() {
	p.lastUpdate = time.Now()
	p.ball.Release()
}
