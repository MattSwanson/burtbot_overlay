package plinko

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/audio"
	"github.com/MattSwanson/ebiten/v2/ebitenutil"
	"github.com/MattSwanson/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	gravity    float64 = 500.0
	gameHeight float64 = 1300
	gameWidth  float64 = 2560
	numRows    int     = 13
	numColumns int     = 19
)

var gameFont font.Face

type Core struct {
	lastUpdate       time.Time
	ball             *ball
	pegs             []*peg
	boxes            []*box
	goalZones        []*zone
	sounds           map[string]*audio.Player
	dropPoints       []fPoint
	currentDropPoint int
	username         string
	rewardMultiplier int
	writeChannel     chan string
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

func Load(screenWidth, screenHeight float64, sounds map[string]*audio.Player, wc chan string) *Core {
	bs, err := os.ReadFile("caskaydia.TTF")
	if err != nil {
		log.Fatal(err)
	}
	tt, err := opentype.Parse(bs)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	gameFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    64,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	ballImg, _, err := ebitenutil.NewImageFromFile("./images/plinko/blue_token.png")
	if err != nil {
		log.Fatal(err)
	}
	pegs := generatePegs(screenWidth, screenHeight)
	dropPoints := []fPoint{}
	for i := 0; i < 5; i++ {
		dropPoints = append(dropPoints, fPoint{(screenWidth/2 - float64(ballImg.Bounds().Dx())) + (float64(i)-2)*300, 20.0})
	}
	b := ball{
		mass:   5,
		x:      dropPoints[2].x,
		y:      dropPoints[2].y,
		img:    ballImg,
		radius: float64(ballImg.Bounds().Dx()) / 2.0,
	}
	boxes := generateBounds(screenWidth, screenHeight)
	zones := generateGoalZones()
	return &Core{
		ball:             &b,
		pegs:             pegs,
		boxes:            boxes,
		sounds:           sounds,
		currentDropPoint: 2,
		dropPoints:       dropPoints,
		goalZones:        zones,
		writeChannel:     wc,
	}
}

func (c *Core) CheckForCollision() {

	if !c.ball.falling {
		return
	}

	const drain float64 = 0.95

	// peg collisions
	for _, peg := range c.pegs {
		dx := c.ball.x - peg.x
		dy := c.ball.y - peg.y
		mag := math.Sqrt(dx*dx + dy*dy)
		vmag := math.Sqrt(c.ball.vx*c.ball.vx + c.ball.vy*c.ball.vy)
		if mag <= peg.radius+c.ball.radius {
			c.ball.vx = (drain * vmag) * (dx / mag)
			c.ball.vy = (drain * vmag) * (dy / mag)
			// .05 -> .25
			//c.sounds["bip"].SetVolume()
			c.sounds["bip"].Rewind()
			c.sounds["bip"].Play()
		}
	}

	// boundary collisions
	for _, box := range c.boxes {
		dY := math.Abs((box.y + 0.5*box.h) - (c.ball.y + c.ball.radius))
		dX := math.Abs((box.x + 0.5*box.w) - (c.ball.x + c.ball.radius))
		if dY < 0.5*box.h+c.ball.radius && dX < box.w/2 {
			c.ball.vy = -c.ball.vy * 0.6
			c.ball.vx = c.ball.vx * 0.6
			c.ball.y = box.y - 2.0*c.ball.radius
			if math.Abs(c.ball.vy) > 25 {
				c.sounds["boing"].Rewind()
				c.sounds["boing"].Play()
			}
		} else if dX < c.ball.radius+0.5*box.w && dY < box.h/2-c.ball.radius {
			c.ball.vx = -c.ball.vx * 0.6
			if c.ball.vx > 0 {
				c.ball.x = box.x + box.w
			} else {
				c.ball.x = box.x - 2.0*c.ball.radius
			}
			c.ball.vy = c.ball.vy * 0.6
			if math.Abs(c.ball.vx) > 25 {
				c.sounds["boing"].Rewind()
				c.sounds["boing"].Play()
			}
		}
	}

	// zone "collisions"
	// all zones min y is 1225.0
	if c.ball.y > 1225 {
		for _, z := range c.goalZones {
			if c.ball.x+c.ball.radius >= z.x && c.ball.x+c.ball.radius < z.x+z.w {
				c.rewardMultiplier = z.rewardValue
				break
			}
		}
	}

	if c.ball.y > gameHeight+200 {
		c.ball.falling = false
		fmt.Printf("you got %d times your token\n", c.rewardMultiplier)
		c.writeChannel <- fmt.Sprintf("plinko result %d\n", c.rewardMultiplier)
	}
}

func (c *Core) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		c.ResetGame(c.username)
		return nil
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) && !c.ball.falling {
		c.MoveDropPoint(c.currentDropPoint - 1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) && !c.ball.falling {
		c.MoveDropPoint(c.currentDropPoint + 1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) && !c.ball.falling {
		c.ReleaseBall()
	}
	c.ball.Update(float64(time.Since(c.lastUpdate).Milliseconds()))
	c.CheckForCollision()
	c.lastUpdate = time.Now()
	return nil
}

func (c *Core) MoveDropPoint(direction int) {
	if direction == -1 && c.currentDropPoint > 0 {
		c.currentDropPoint--
		c.ball.x, c.ball.y = c.dropPoints[c.currentDropPoint].x, c.dropPoints[c.currentDropPoint].y
	} else if direction == 1 && c.currentDropPoint < len(c.dropPoints)-1 {
		c.currentDropPoint++
		c.ball.x, c.ball.y = c.dropPoints[c.currentDropPoint].x, c.dropPoints[c.currentDropPoint].y
	}
}

func (c *Core) Draw(screen *ebiten.Image) {
	c.ball.Draw(screen)
	for _, v := range c.pegs {
		v.Draw(screen)
	}
	for _, v := range c.boxes {
		v.Draw(screen)
	}
	for _, v := range c.goalZones {
		v.Draw(screen)
	}
	text.Draw(screen, "Current Player: "+c.username, gameFont, 30, 50, color.RGBA{0x00, 0xff, 0x00, 0xff})
	cx, cy := ebiten.CursorPosition()
	ebitenutil.DebugPrint(screen, fmt.Sprintf("cx: %d, cy: %d", cx, cy))
}

func (c *Core) ResetGame(username string) {
	c.ball.falling = false
	c.ball.vx = 0
	c.ball.vy = 0
	c.currentDropPoint = 2
	c.ball.x = c.dropPoints[c.currentDropPoint].x
	c.ball.y = c.dropPoints[c.currentDropPoint].y
	c.username = username
}

func (c *Core) ReleaseBall() {
	c.lastUpdate = time.Now()
	c.ball.Release()
}

func generateBounds(screenWidth, screenHeight float64) []*box {
	boxes := make([]*box, 2)

	boxHeight := gameHeight
	boxWidth := 0.025 * gameWidth
	x := 0.5 * (screenWidth - gameWidth)
	y := gameHeight - boxHeight + 0.5*(screenHeight-gameHeight)
	bounds := fRect{fPoint{x, y}, fPoint{x + boxWidth, y + boxHeight}}
	boxtwo := NewBox(bounds, color.RGBA{0x00, 0x00, 0xff, 0xff})
	boxes[0] = boxtwo

	boxHeight = gameHeight
	boxWidth = 0.025 * gameWidth

	x = screenWidth - boxWidth - 0.5*(screenWidth-gameWidth)
	y = gameHeight - boxHeight + 0.5*(screenHeight-gameHeight)
	bounds = fRect{fPoint{x, y}, fPoint{x + boxWidth, y + boxHeight}}
	boxthree := NewBox(bounds, color.RGBA{0x00, 0x00, 0xff, 0xff})
	boxes[1] = boxthree

	return boxes
}

func generatePegs(screenWidth, screenHeight float64) []*peg {
	pegImg, _, err := ebitenutil.NewImageFromFile("./images/plinko/ball.png")
	if err != nil {
		log.Fatal(err)
	}
	pegs := make([]*peg, numColumns*numRows)
	halfImgWidth := float64(pegImg.Bounds().Dx()) / 2.0
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
	return pegs
}

func generateGoalZones() []*zone {
	zoneCount := 11
	zones := make([]*zone, zoneCount)
	for i := 0; i < zoneCount; i++ {
		w := gameWidth / float64(zoneCount)
		min := fPoint{
			x: float64(i) * w,
			y: 1225.0,
		}
		max := fPoint{
			x: min.x + w,
			y: gameHeight,
		}
		z := NewZone(fRect{min, max}, rand.Intn(10)+1)
		zones[i] = z
	}
	return zones
}
