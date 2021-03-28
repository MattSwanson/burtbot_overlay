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
var playerLabelFont font.Face
var tokenImg *ebiten.Image

type Core struct {
	lastUpdate       time.Time
	tokens           []*token
	pegs             []*peg
	boxes            []*box
	goalZones        []*zone
	sounds           map[string]*audio.Player
	dropPoints       []fPoint
	currentDropPoint int
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

	playerLabelFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	tokenImg, _, err = ebitenutil.NewImageFromFile("./images/plinko/blue_token.png")
	if err != nil {
		log.Fatal(err)
	}
	pegs := generatePegs(screenWidth, screenHeight)
	dropPoints := []fPoint{}
	for i := 0; i < 5; i++ {
		dropPoints = append(dropPoints, fPoint{(screenWidth/2 - float64(tokenImg.Bounds().Dx())) + (float64(i)-2)*300, 20.0})
	}

	tokens := []*token{}
	boxes := generateBounds(screenWidth, screenHeight)
	zones := generateGoalZones()
	return &Core{
		tokens:           tokens,
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

	const drain float64 = 0.95
	for idx, b := range c.tokens {
		if !b.falling {
			continue
		}
		// peg collisions
		for _, peg := range c.pegs {
			dx := b.x - peg.x
			dy := b.y - peg.y
			mag := math.Sqrt(dx*dx + dy*dy)
			vmag := math.Sqrt(b.vx*b.vx + b.vy*b.vy)
			if mag <= peg.radius+b.radius {
				b.vx = (drain * vmag) * (dx / mag)
				b.vy = (drain * vmag) * (dy / mag)
				// .05 -> .25
				//c.sounds["bip"].SetVolume()
				//c.sounds["bip"].Rewind()
				//c.sounds["bip"].Play()
			}
		}

		// token collisions????
		for otidx, ot := range c.tokens {
			if otidx == idx {
				continue
			}
			dx := b.x - ot.x
			dy := b.y - ot.y

			// magnitude of the collsion vector
			mag := math.Sqrt(dx*dx + dy*dy)

			if mag <= ot.radius+b.radius {
				// magnitude of this tokens velocity
				vmag := math.Sqrt(b.vx*b.vx + b.vy*b.vy)

				// magnitude of other tokens velocity
				otvmag := math.Sqrt(ot.vx*ot.vx + ot.vy*ot.vy)

				// total velocity of the collision -- masses are equal so no need to worky about that
				totalVelocity := vmag + otvmag

				b.vx = (drain * totalVelocity) / 2.0 * (dx / mag)
				b.vy = (drain * totalVelocity) / 2.0 * (dy / mag)
				ot.vx = (drain * totalVelocity) / 2.0 * (-1.0 * dx / mag)
				ot.vy = (drain * totalVelocity) / 2.0 * (-1.0 * dy / mag)
				c.sounds["boing"].Rewind()
				c.sounds["boing"].Play()
			}
		}

		// boundary collisions
		for _, box := range c.boxes {
			dY := math.Abs((box.y + 0.5*box.h) - (b.y + b.radius))
			dX := math.Abs((box.x + 0.5*box.w) - (b.x + b.radius))
			if dY < 0.5*box.h+b.radius && dX < box.w/2 {
				b.vy = -b.vy * 0.6
				b.vx = b.vx * 0.6
				b.y = box.y - 2.0*b.radius
				if math.Abs(b.vy) > 25 {
					c.sounds["boing"].Rewind()
					c.sounds["boing"].Play()
				}
			} else if dX < b.radius+0.5*box.w && dY < box.h/2-b.radius {
				b.vx = -b.vx * 0.6
				if b.vx > 0 {
					b.x = box.x + box.w
				} else {
					b.x = box.x - 2.0*b.radius
				}
				b.vy = b.vy * 0.6
				if math.Abs(b.vx) > 25 {
					c.sounds["boing"].Rewind()
					c.sounds["boing"].Play()
				}
			}
		}

		// zone "collisions"
		// all zones min y is 1225.0
		if b.y > 1225 {
			for _, z := range c.goalZones {
				if b.x+b.radius >= z.x && b.x+b.radius < z.x+z.w {
					c.rewardMultiplier = z.rewardValue
					break
				}
			}
		}

		if b.y > gameHeight+200 {
			b.falling = false
			fmt.Printf("you got %d times your token\n", c.rewardMultiplier)
			c.writeChannel <- fmt.Sprintf("plinko result %s %d\n", b.playerName, c.rewardMultiplier)
			c.tokens = removeBall(c.tokens, idx)
		}
	}
}

func removeBall(s []*token, i int) []*token {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func (c *Core) Update() error {
	delta := float64(time.Since(c.lastUpdate).Milliseconds())
	for _, b := range c.tokens {
		b.Update(delta)
	}
	c.CheckForCollision()
	c.lastUpdate = time.Now()
	return nil
}

func (c *Core) Draw(screen *ebiten.Image) {
	for _, v := range c.tokens {
		v.Draw(screen)
	}
	for _, v := range c.pegs {
		v.Draw(screen)
	}
	for _, v := range c.boxes {
		v.Draw(screen)
	}
	for _, v := range c.goalZones {
		v.Draw(screen)
	}
	for k, v := range c.dropPoints {
		text.Draw(screen, fmt.Sprint(k), gameFont, int(v.x), int(v.y)+35, color.RGBA{0x00, 0xff, 0x00, 0xff})
	}
}

func (c *Core) DropBall(pos int, playerName string) {
	// make a new token with its pos set to the selected drop point
	if pos < 0 || pos > len(c.dropPoints) {
		return // do nothing for now, but should return an error
		// or is this validated on the other end? no
	}
	b := NewBall(playerName, tokenImg, c.dropPoints[pos])
	c.tokens = append(c.tokens, b)
	b.Release()
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
	pegImg, _, err := ebitenutil.NewImageFromFile("./images/plinko/token.png")
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
