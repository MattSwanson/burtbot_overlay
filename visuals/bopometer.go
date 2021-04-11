package visuals

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/ebitenutil"
	"github.com/MattSwanson/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

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
	bopFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    96,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	bigBopFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    256,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := ebitenutil.NewImageFromFile("./images/bopM.png")
	if err != nil {
		log.Fatal(err)
	}
	mediumBop = img

	img, _, err = ebitenutil.NewImageFromFile("./images/bopL.png")
	if err != nil {
		log.Fatal(err)
	}
	largeBop = img

	img, _, err = ebitenutil.NewImageFromFile("./images/bopometer_bg.png")
	if err != nil {
		log.Fatal(err)
	}
	bg = img
}

const (
	gravity    = 700
	finalLabel = "Final Rating:"
)

var bopFont font.Face
var bigBopFont font.Face
var mediumBop *ebiten.Image
var largeBop *ebiten.Image
var bg *ebiten.Image
var finalLabelX int

type Bopometer struct {
	currentRating float64
	totalBops     int
	running       bool
	finished      bool
	bops          []*bop
	bopIndicatorY float64
	writeChannel  chan string
}

func NewBopometer(wc chan string) *Bopometer {
	finalLabelX = int(text.BoundString(bopFont, finalLabel).Dx() / 2)
	return &Bopometer{bops: []*bop{}, writeChannel: wc}
}

func (b *Bopometer) Draw(screen *ebiten.Image) {
	const textYOffset = 35.0
	const textX = 200.0
	const bopIndicatorX = 135.0
	if b.running {
		screen.DrawImage(bg, nil)
		op := ebiten.DrawImageOptions{}
		op.GeoM.Translate(-float64(largeBop.Bounds().Dx())/2, -float64(largeBop.Bounds().Dy())/2)
		op.GeoM.Rotate(math.Pi/2 + 0.1)
		op.GeoM.Translate(bopIndicatorX, b.bopIndicatorY)
		screen.DrawImage(largeBop, &op)
		op.GeoM.Reset()
		text.Draw(screen, fmt.Sprintf("%.2f", b.currentRating), bopFont, textX, int(b.bopIndicatorY)+textYOffset, color.RGBA{0xff, 0x00, 0x00, 0xff})
		for _, bp := range b.bops {
			bp.Draw(screen)
		}
	}
	if b.finished {
		text.Draw(screen, finalLabel, bopFont, finalLabelX, 400, color.RGBA{0xff, 0x00, 0x00, 0xff})
		text.Draw(screen, fmt.Sprintf("%.2f", b.currentRating), bigBopFont, 800, 720, color.RGBA{0xff, 0x00, 0x00, 0xff})
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
	img *ebiten.Image
}

func (b *bop) Update(delta float64) {
	b.x += b.vx * delta / 1000
	b.vy = b.vy + gravity*delta/1000.0
	b.y += b.vy * delta / 1000
	b.a += b.va * delta / 1000
}

func (b *bop) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Rotate(b.a)
	op.GeoM.Translate(b.x, b.y)
	screen.DrawImage(b.img, &op)
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
func calculateRating(numBops int) float64 {
	switch {
	case numBops == 0:
		return 0.0
	case numBops == 1:
		return 1.0
	case numBops < 3:
		return 2.0
	case numBops < 7:
		return 3.0 + (float64(numBops)-3.0)/4.0
	case numBops < 11:
		return 4.0 + (float64(numBops)-7.0)/4.0
	case numBops < 100:
		return 5.0 + (float64(numBops)-11.0)/89.0
	case numBops < 1000:
		return 6.0 + (float64(numBops)-100.0)/900.0
	case numBops < 10000:
		return 7.0 + (float64(numBops)-1000.0)/9000.0
	case numBops < 100000:
		return 8.0 + (float64(numBops)-10000.0)/90000.0
	case numBops < 1000000:
		return 9.0 + (float64(numBops)-100000.0)/900000.0
	default:
		return 10.0
	}
}
