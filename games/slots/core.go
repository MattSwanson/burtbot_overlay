package slots

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"math/rand"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

const (
	drawOffsetX       float32       = 128.0
	drawOffsetY       float32       = 128.0
	spinVelocity      float32       = 1536.0
	reelOneSpinTime   time.Duration = 5_000  // ms
	reelTwoSpinTime   time.Duration = 7_500  // ms
	reelThreeSpinTime time.Duration = 10_000 // ms
)

var sevenImg *rl.Image
var barImg *rl.Image
var bellImg *rl.Image
var cherryImg *rl.Image
var coconutImg *rl.Image
var pearImg *rl.Image
var watermelonImg *rl.Image

type reel struct {
	texture       rl.Texture2D
	frameStartY   float32
	currentSymbol int
	targetSymbol  int
	symbolOrder   []int
	isSpinning    bool
	velocity      float32
	// 0 cherry - 1 watermelon - 2 pear - 3 coconut - 4 bell - 5 bar - 6 seven
}

type Core struct {
	currentBet   int
	currentUser  string
	lastUpdate   time.Time
	isActive     bool
	writeChannel chan string
	reels        []*reel
}

func LoadSlots(wc chan string) *Core {

	cherryImg = rl.LoadImage("./images/slots/cherry.png")
	coconutImg = rl.LoadImage("./images/slots/coconut.png")
	barImg = rl.LoadImage("./images/slots/bar.png")
	bellImg = rl.LoadImage("./images/slots/bell.png")
	pearImg = rl.LoadImage("./images/slots/pear.png")
	watermelonImg = rl.LoadImage("./images/slots/watermelon.png")
	sevenImg = rl.LoadImage("./images/slots/seven.png")

	reels := []*reel{
		newReel(),
		newReel(),
		newReel(),
	}
	c := Core{
		writeChannel: wc,
		reels:        reels,
	}
	return &c
}

func newReel() *reel {
	nums := []int{0, 1, 2, 3, 4, 5, 6}
	rand.Shuffle(len(nums), func(i, j int) {
		nums[i], nums[j] = nums[j], nums[i]
	})
	texture := generateReelTexture(nums)
	return &reel{
		texture:     texture,
		frameStartY: -64,
		symbolOrder: nums,
		velocity:    spinVelocity,
	}
}

// Create a composite reel texure using the order of symbols specified
// by the int slice given
func generateReelTexture(order []int) rl.Texture2D {
	buf := []byte{}

	// cherry img first
	for _, v := range order {
		var symImage *rl.Image
		switch v {
		case 0:
			symImage = cherryImg
		case 1:
			symImage = watermelonImg
		case 2:
			symImage = pearImg
		case 3:
			symImage = coconutImg
		case 4:
			symImage = bellImg
		case 5:
			symImage = barImg
		case 6:
			symImage = sevenImg
		}
		buf = append(buf, getRlImageBytes(symImage)...)
	}

	compImg := rl.NewImage(buf, 256, int32(len(order)*128), 1, rl.UncompressedR8g8b8a8)
	return rl.LoadTextureFromImage(compImg)
}

// get the color data from an image, rl.UncompressedR8g8b8a8
func getRlImageBytes(img *rl.Image) []byte {
	buf := []byte{}
	cImg := img.ToImage()
	for y := 0; y < 128; y++ {
		for x := 0; x < 256; x++ {
			r, g, b, a := cImg.At(x, y).RGBA()
			buf = append(buf, byte(r))
			buf = append(buf, byte(g))
			buf = append(buf, byte(b))
			buf = append(buf, byte(a))
		}
	}
	return buf
}

func (c *Core) Update(d float64) {
	for i := range c.reels {
		if !c.reels[i].isSpinning {
			continue
		}
		c.reels[i].frameStartY -= c.reels[i].velocity * float32(d) / 1000.0
	}
	c.lastUpdate = time.Now()
}

func ScoreReels(reels []*reel) float64 {
	mult := 0.0
	if reels[0].currentSymbol == reels[1].currentSymbol &&
		reels[0].currentSymbol == reels[2].currentSymbol {
		switch reels[0].currentSymbol {
		case 0:
			mult = 4.0
		case 1:
			mult = 6.0
		case 2:
			mult = 8.0
		case 3:
			mult = 10.0
		case 4:
			mult = 30.0
		case 5:
			mult = 50.0
		case 6:
			mult = 80.0
		}
	} else if reels[0].currentSymbol == reels[1].currentSymbol {
		switch reels[0].currentSymbol {
		case 0:
			mult = 0.6
		case 6:
			if reels[2].currentSymbol == 5 {
				mult = 3.0
			}
		}
	} else {
		switch reels[0].currentSymbol {
		case 0:
			mult = 0.2
		case 5:
			if reels[1].currentSymbol == reels[2].currentSymbol {
				switch reels[1].currentSymbol {
				case 1:
					mult = 1.0
				case 2:
					mult = 1.4
				case 3:
					mult = 1.8
				case 4:
					mult = 2.2
				}
			}
		case 6:
			if reels[1].currentSymbol == 5 &&
				reels[2].currentSymbol == 5 {
				mult = 2.6
			}
		}
	}
	return mult
}

func (c *Core) reset() {
	for i := range c.reels {
		c.reels[i].isSpinning = false
		c.reels[i].frameStartY = -64
	}
}

func (c *Core) HandleMessage(args []string) {
	switch args[0] {
	case "start":
		c.isActive = true
	case "pull":
		c.Pull(args)
	case "stop":
		c.isActive = false
		c.reset()
	}
}

func (c *Core) Pull(args []string) {
	if len(args) < 3 {
		return
	}
	bet, err := strconv.Atoi(args[1])
	if err != nil || bet <= 0 {
		return
	}
	c.currentUser = args[2]
	c.currentBet = bet
	c.isActive = true
	for i := range c.reels {
		c.reels[i].targetSymbol = rand.Intn(7)
		c.reels[i].isSpinning = true
	}
	go func() {
		rm := time.Duration(500 - rand.Float64()*1000)
		time.Sleep((reelOneSpinTime + rm) * time.Millisecond)
		c.reels[0].isSpinning = false
		idx := (7 - (int(c.reels[0].frameStartY)%896)/-128) % 7
		c.reels[0].frameStartY = float32(idx*128 - 64)
		c.reels[0].currentSymbol = c.reels[0].symbolOrder[idx]
	}()
	go func() {
		rm := time.Duration(500 - rand.Float64()*1000)
		time.Sleep((reelTwoSpinTime + rm) * time.Millisecond)
		c.reels[1].isSpinning = false
		idx := (7 - (int(c.reels[0].frameStartY)%896)/-128) % 7
		c.reels[1].frameStartY = float32(idx*128 - 64)
		c.reels[1].currentSymbol = c.reels[1].symbolOrder[idx]
	}()
	go func() {
		rm := time.Duration(500 - rand.Float64()*1000)
		time.Sleep((reelThreeSpinTime + rm) * time.Millisecond)
		c.reels[2].isSpinning = false
		idx := (7 - (int(c.reels[0].frameStartY)%896)/-128) % 7
		c.reels[2].frameStartY = float32(idx*128 - 64)
		c.reels[2].currentSymbol = c.reels[2].symbolOrder[idx]
		mult := ScoreReels(c.reels)
		payout := int(math.Ceil(mult * float64(c.currentBet)))
		time.Sleep(5 * time.Second)
		c.writeChannel <- fmt.Sprintf("slots result %s %d\n", c.currentUser, payout)
		c.isActive = false
		c.reset()
	}()
}

func (c *Core) Draw() {
	if !c.isActive {
		return
	}
	for i, reel := range c.reels {
		rl.DrawTexturePro(
			reel.texture,
			rl.Rectangle{X: 0, Y: reel.frameStartY, Width: 256, Height: 256},
			rl.Rectangle{
				X:      drawOffsetX + 25.0 + float32(i)*256.0,
				Y:      drawOffsetY + 0,
				Width:  256,
				Height: 256},
			rl.Vector2{X: 0, Y: 0},
			0.0,
			rl.White,
		)
	}
}

func (c *Core) Cleanup() {

}
