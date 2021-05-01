package visuals

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

const (
	gravity       = 700
	textSize      = 96
	largeTextSize = 256
	finalLabel    = "Final Rating:"
)

var bopFont rl.Font
var mediumBop rl.Texture2D
var largeBop rl.Texture2D
var bg rl.Texture2D
var finalLabelX int

type Bopometer struct {
	currentRating float32
	totalBops     int
	running       bool
	finished      bool
	bops          []*bop
	bopIndicatorY float32
	writeChannel  chan string
}

func LoadBopometerAssets() {
	mediumBop = rl.LoadTexture("./images/bopM.png")
	largeBop = rl.LoadTexture("./images/bopL.png")
	bg = rl.LoadTexture("./images/bopometer_bg.png")
	bopFont = rl.LoadFont("caskaydia.TTF")
}

func NewBopometer(wc chan string) *Bopometer {
	finalLabelX = int(rl.MeasureTextEx(bopFont, finalLabel, 96, 0).X / 2)
	return &Bopometer{bops: []*bop{}, writeChannel: wc}
}

func (b *Bopometer) Draw() {
	const textYOffset = 35.0
	const textX = 200.0
	const bopIndicatorX = 200.0
	if b.running {
		rl.DrawTexture(bg, 0, 0, rl.White)
		rl.DrawTextureEx(largeBop, rl.Vector2{X: bopIndicatorX, Y: b.bopIndicatorY}, 90, 1, rl.White)
		txtPos := rl.Vector2{X: textX, Y: b.bopIndicatorY + textYOffset}
		rl.DrawTextEx(bopFont, fmt.Sprintf("%.2f", b.currentRating), txtPos, textSize, 0, rl.Red)
		for _, bp := range b.bops {
			bp.Draw()
		}
	}
	if b.finished {
		rl.DrawTextEx(bopFont, finalLabel, rl.Vector2{X: float32(finalLabelX), Y: 400}, largeTextSize, 0, rl.Red)
		rl.DrawTextEx(bopFont, fmt.Sprintf("%.2f", b.currentRating), rl.Vector2{X: 800, Y: 720}, largeTextSize, 0, rl.Red)
	}
}

func (b *Bopometer) Update(delta float64) error {
	b.bopIndicatorY = 1400 - b.currentRating/10.0*1350
	for k, bp := range b.bops {
		bp.Update(delta)
		if bp.x > 2700 {
			removeBop(b.bops, k)
		}
	}
	return nil
}

func removeBop(s []*bop, i int) []*bop {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func (b *Bopometer) Add(n int) {
	b.totalBops += n
	b.currentRating = calculateRating(b.totalBops)
	spawnLeft := rand.Int63()%2 == 0
	x, y := 2585.0, 1465.0
	if spawnLeft {
		x, y = -25, 1465
	}
	for i := 0; i < n; i++ {
		newbop := spawnBop()
		newbop.SetPosition(x, y)
		rvx := -1 * (rand.Intn(600) + 600)
		rvy := -1 * (rand.Intn(800) + 800)
		rva := -1 * (rand.Float64() - 0.5 + math.Pi)
		if spawnLeft {
			rvx *= -1
			rva *= -1
		}
		newbop.SetVelocity(float64(rvx), float64(rvy), rva)
		b.bops = append(b.bops, newbop)
	}
}

func (b *Bopometer) SetRunning(bo bool) { b.running = bo }
func (b *Bopometer) IsRunning() bool    { return b.running }
func (b *Bopometer) IsFinished() bool   { return b.finished }
func (b *Bopometer) Reset()             { b.currentRating = 0; b.bops = []*bop{}; b.totalBops = 0 }
func (b *Bopometer) Finish() {
	b.writeChannel <- fmt.Sprintf("bop result %.2f\n", b.currentRating)
	b.finished = true
	go func() {
		time.Sleep(time.Second * 10)
		b.finished = false
	}()
}

type bop struct {
	x   float64
	y   float64
	vx  float64
	vy  float64
	a   float64
	va  float64
	img rl.Texture2D
}

func (b *bop) Update(delta float64) {
	b.x += b.vx * delta / 1000
	b.vy = b.vy + gravity*delta/1000.0
	b.y += b.vy * delta / 1000
	b.a += b.va * delta / 1000
}

func (b *bop) Draw() {
	rl.DrawTextureEx(b.img, rl.Vector2{X: float32(b.x), Y: float32(b.y)}, float32(b.a)*180/math.Pi, 1, rl.White)
}

func (b *bop) SetVelocity(x, y, a float64) {
	b.vx, b.vy, b.va = x, y, a
}

func (b *bop) SetPosition(x, y float64) {
	b.x, b.y = x, y
}

func spawnBop() *bop {
	return &bop{img: mediumBop}
}

// calculateRating will translate the number of bops given
// to its bopometer rating
func calculateRating(numBops int) float32 {
	switch {
	case numBops == 0:
		return 0.0
	case numBops == 1:
		return 1.0
	case numBops < 3:
		return 2.0
	case numBops < 7:
		return 3.0 + (float32(numBops)-3.0)/4.0
	case numBops < 11:
		return 4.0 + (float32(numBops)-7.0)/4.0
	case numBops < 100:
		return 5.0 + (float32(numBops)-11.0)/89.0
	case numBops < 1000:
		return 6.0 + (float32(numBops)-100.0)/900.0
	case numBops < 10000:
		return 7.0 + (float32(numBops)-1000.0)/9000.0
	case numBops < 100000:
		return 8.0 + (float32(numBops)-10000.0)/90000.0
	case numBops < 1000000:
		return 9.0 + (float32(numBops)-100000.0)/900000.0
	default:
		return 10.0
	}
}
