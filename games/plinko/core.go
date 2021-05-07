package plinko

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/MattSwanson/raylib-go/physics"
	rl "github.com/MattSwanson/raylib-go/raylib"
)

const (
	gravity       float64 = 500.0
	gameHeight    float64 = 1440
	gameWidth     float64 = 2560
	numRows       int     = 13
	numColumns    int     = 25
	numDropQueues int     = 5
)

var gameFont font.Face
var playerLabelFont font.Face
var tokenImg rl.Texture2D
var barrierImg rl.Texture2D
var timerChannel chan bool

var goalValues = []int{1, 0, 2, 1, 0, 5, 0, 1, 2, 0, 1}

type Core struct {
	lastUpdate       time.Time
	tokens           []*token
	pegs             []*peg
	goalZones        []*zone
	barriers         []*barrier
	queues           []tokenQueue
	sounds           map[string]rl.Sound
	currentDropPoint int
	rewardMultiplier int
	writeChannel     chan string
	CancelTimer      context.CancelFunc
}

type fPoint struct {
	x float64
	y float64
}

type fRect struct {
	min fPoint
	max fPoint
}

type vec2f struct {
	x float64
	y float64
}

func dot(a, b vec2f) float64 {
	return a.x*b.x + a.y*b.y
}

func add(a, b vec2f) vec2f {
	return vec2f{a.x + b.x, a.y + b.y}
}

func scale(v vec2f, s float64) vec2f {
	return vec2f{v.x * s, v.y * s}
}

func sub(a, b vec2f) vec2f {
	return vec2f{a.x - b.x, a.y - b.y}
}

func angle(a, b vec2f) float64 {
	return math.Acos(dot(a, b) / mag(a) * mag(b))
}

func mag(a vec2f) float64 {
	return math.Sqrt(a.x*a.x + a.y*a.y)
}

// reflect will return a vector created by
// reflected input a across normal vector n
// maybe this should just normalize the second
// arg?
func reflect(a, n vec2f) vec2f {
	// 2(a + n(-a dot n)) - a
	v := scale(a, -1)
	v = scale(n, dot(v, n))
	v = add(a, v)
	v = scale(v, 2)
	v = sub(v, a)
	return v
}

func (r *fRect) Dx() float64 {
	return r.max.x - r.min.x
}

func (r *fRect) Dy() float64 {
	return r.max.y - r.min.y
}

type tokenQueue struct {
	Tokens       []*token
	dropPosition fPoint
}

// push the token to the back of the queue
func (tq *tokenQueue) push(t *token) {
	tq.Tokens = append(tq.Tokens, t)
}

// pop the front element from the front of the queue
func (tq *tokenQueue) pop() (*token, error) {
	if len(tq.Tokens) == 0 {
		return nil, errors.New("nothing in queue")
	}
	t := tq.Tokens[0]
	if len(tq.Tokens) == 1 {
		tq.Tokens = []*token{}
	} else {
		tq.Tokens = tq.Tokens[1:]
	}
	return t, nil
}

func Load(screenWidth, screenHeight float64, wc chan string, sounds map[string]rl.Sound) *Core {
	timerChannel = make(chan bool)
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

	tokenImg = rl.LoadTexture("./images/plinko/white_token.png")
	if err != nil {
		log.Fatal(err)
	}

	barrierImg = rl.LoadTexture("./images/plinko/triangle.png")
	if err != nil {
		log.Fatal(err)
	}

	pegs := generatePegs(screenWidth, screenHeight)
	//dropPoints := []fPoint{}
	tokenQueues := make([]tokenQueue, numDropQueues)
	for i := 0; i < numDropQueues; i++ {
		dropPoint := fPoint{(screenWidth/2 - float64(tokenImg.Width)) + float64(i-numDropQueues/2)*300, 20.0}
		tq := tokenQueue{
			Tokens:       []*token{},
			dropPosition: dropPoint,
		}
		tokenQueues[i] = tq
	}

	tokens := []*token{}
	//boxes := generateBounds(screenWidth, screenHeight)
	zones := generateGoalZones()
	barriers := generateBarriers(len(zones) + 1)

	c := Core{
		tokens:           tokens,
		pegs:             pegs,
		sounds:           sounds,
		currentDropPoint: 2,
		queues:           tokenQueues,
		barriers:         barriers,
		goalZones:        zones,
		writeChannel:     wc,
	}
	c.CancelTimer = manageQueues()
	return &c
}

func (c *Core) CheckForCollision(delta float64) {

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

				rl.PlaySoundMulti(c.sounds["boing"])
			}
		}

		if b.x <= 0 {
			b.vy = b.vy * 0.6
			b.vx = -b.vx * 0.6
			b.x = 1
		}
		if b.x+2*b.radius >= gameWidth {
			b.vy = b.vy * 0.6
			b.vx = -b.vx * 0.6
			b.x = gameWidth - 2*b.radius - 1
		}

		// barrier collision
		for _, barrier := range c.barriers {
			if b.y < 1300 {
				break
			}
			dx := b.x + b.radius - barrier.x

			// which barrier are we closest to?

			if dx > 0 {
				px := b.x + b.radius + b.radius*math.Cos(3.926991)
				py := b.y + b.radius - b.radius*math.Sin(3.926991)
				if barrier.bounds[0].IsLeft(px, py) <= 0 &&
					barrier.bounds[1].IsLeft(px, py) <= 0 &&
					barrier.bounds[2].IsLeft(px, py) <= 0 {

					nx := math.Cos(math.Pi / 4.0)
					ny := -math.Sin(math.Pi / 4.0)

					bbx, bby := barrier.bounds[0].x0, barrier.bounds[0].y0
					nOffset := rl.Vector2DotProduct(rl.Vector2{X: float32(nx), Y: float32(ny)}, rl.Vector2{X: float32(bbx), Y: float32(bby)})
					dist := rl.Vector2DotProduct(rl.Vector2{X: float32(nx), Y: float32(ny)}, rl.Vector2{X: float32(px), Y: float32(py)})
					dist -= nOffset
					r := reflect(vec2f{b.vx, b.vy}, vec2f{nx, ny})
					r = scale(r, 0.6)
					b.vx, b.vy = r.x, r.y
					b.x += float64(-dist) * math.Cos(nx)
					b.y -= float64(-dist) * -math.Sin(ny)
				}
			} else if dx < 0 {
				px := b.x + b.radius + b.radius*math.Cos(5.497787)
				py := b.y + b.radius - b.radius*math.Sin(5.497787)
				if barrier.bounds[0].IsLeft(px, px) <= 0 &&
					barrier.bounds[1].IsLeft(px, py) <= 0 &&
					barrier.bounds[2].IsLeft(px, py) <= 0 {

					nx := math.Cos(3 * math.Pi / 4.0)
					ny := -math.Sin(3 * math.Pi / 4.0)

					bbx, bby := barrier.bounds[1].x0, barrier.bounds[1].y0
					nOffset := rl.Vector2DotProduct(rl.Vector2{X: float32(nx), Y: float32(ny)}, rl.Vector2{X: float32(bbx), Y: float32(bby)})
					dist := rl.Vector2DotProduct(rl.Vector2{X: float32(nx), Y: float32(ny)}, rl.Vector2{X: float32(px), Y: float32(py)})
					dist -= nOffset

					r := reflect(vec2f{b.vx, b.vy}, vec2f{nx, ny})
					r = scale(r, 0.6)
					b.vx, b.vy = r.x, r.y
					b.x -= float64(-dist) * math.Cos(nx)
					b.y -= float64(-dist) * -math.Sin(ny)

					fmt.Println(b.vx, b.vy)
				}
			} else {
				px := b.x + b.radius
				py := b.y + 2*b.radius
				if barrier.bounds[0].IsLeft(px, py) <= 0 &&
					barrier.bounds[1].IsLeft(px, py) <= 0 &&
					barrier.bounds[2].IsLeft(px, py) <= 0 {
					b.vy *= -0.6
					b.vx *= 0.6
				}
			}
		}

		// boundary collisions
		// for _, box := range c.boxes {
		// 	dY := math.Abs((box.y + 0.5*box.h) - (b.y + b.radius))
		// 	dX := math.Abs((box.x + 0.5*box.w) - (b.x + b.radius))
		// 	if dY < 0.5*box.h+b.radius && dX < box.w/2 {
		// 		b.vy = -b.vy * 0.6
		// 		b.vx = b.vx * 0.6
		// 		b.y = box.y - 2.0*b.radius
		// 		if math.Abs(b.vy) > 25 {
		// 			c.sounds["boing"].Rewind()
		// 			c.sounds["boing"].Play()
		// 		}
		// 	} else if dX < b.radius+0.5*box.w && dY < box.h/2-b.radius {
		// 		b.vx = -b.vx * 0.6
		// 		if b.vx > 0 {
		// 			b.x = box.x + box.w
		// 		} else {
		// 			b.x = box.x - 2.0*b.radius
		// 		}
		// 		b.vy = b.vy * 0.6
		// 		if math.Abs(b.vx) > 25 {
		// 			c.sounds["boing"].Rewind()
		// 			c.sounds["boing"].Play()
		// 		}
		// 	}
		// }

		// zone "collisions"
		// all zones min y is 1225.0
		if b.y > 1400 {
			for _, z := range c.goalZones {
				if b.x+b.radius >= z.x && b.x+b.radius < z.x+z.w {
					c.rewardMultiplier = z.rewardValue
					z.Hit()
					break
				}
			}
		}

		if b.y > gameHeight+50 {
			b.falling = false
			fmt.Printf("you got %d times your token\n", c.rewardMultiplier)
			c.writeChannel <- fmt.Sprintf("plinko result %s %d\n", b.playerName, c.rewardMultiplier)
			physics.DestroyBody(c.tokens[idx].physBody)
			c.tokens = removeBall(c.tokens, idx)
		}
	}
}

func removeBall(s []*token, i int) []*token {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func (c *Core) Update() error {
	select {
	case <-timerChannel:
		for i := 0; i < len(c.queues); i++ {
			t, err := c.queues[i].pop()
			if err != nil {
				continue
			}
			c.tokens = append(c.tokens, t)
			t.SetPosition(c.queues[i].dropPosition.x, c.queues[i].dropPosition.y)
			t.Release()
		}
	default:
	}

	delta := float64(time.Since(c.lastUpdate).Milliseconds())

	for _, b := range c.tokens {
		b.Update(delta)
	}
	c.CheckForCollision(delta)
	c.lastUpdate = time.Now()

	return nil
}

func (c *Core) Draw() {
	if len(c.tokens) == 0 {
		return
	}
	for _, v := range c.tokens {
		v.Draw()
	}
	for _, v := range c.pegs {
		v.Draw()
	}
	for _, v := range c.barriers {
		v.Draw()
	}
	for _, v := range c.goalZones {
		v.Draw()
	}
	for k, v := range c.queues {
		//text.Draw(screen, , gameFont, int(v.dropPosition.x), int(v.dropPosition.y)+35, color.RGBA{0x00, 0xff, 0x00, 0xff})
		rl.DrawText(fmt.Sprint(k), int32(v.dropPosition.x), int32(v.dropPosition.y)+35, 72, rl.Green)
	}
}

func (c *Core) DropBall(pos int, playerName, playerColor string) {
	// make a new token with its pos set to the selected drop point
	if pos < 0 || pos >= len(c.queues) {
		return // do nothing for now, but should return an error
		// or is this validated on the other end? no
	}
	t := NewToken(playerName, playerColor, tokenImg, c.queues[pos].dropPosition)
	c.queues[pos].push(t)
}

func (c *Core) DropAll(playerName, playerColor string) {
	for i := 0; i < len(c.queues); i++ {
		c.DropBall(i, playerName, playerColor)
	}
}

func manageQueues() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				timerChannel <- true
			}
		}
	}(ctx)
	return cancel
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
	pegImg := rl.LoadTexture("./images/plinko/token.png")
	pegs := make([]*peg, numColumns*numRows)
	halfImgWidth := float64(pegImg.Width) / 2.0
	offset := 25.0
	for i := 0; i < len(pegs); i++ {
		if i%numColumns == 0 {
			offset *= -1
		}
		x := (float64((i%numColumns)-numColumns/2)*100.0 + (screenWidth/2 - halfImgWidth)) + offset
		y := float64(i/numColumns)*75.0 + 200.0
		radius := float64(pegImg.Width) / 2.0
		p := peg{
			x:      x,
			y:      y,
			radius: radius,
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
			y: gameHeight,
		}
		max := fPoint{
			x: min.x + w,
			y: gameHeight + 10,
		}
		z := NewZone(fRect{min, max}, goalValues[i])
		zones[i] = z
	}
	return zones
}

func generateBarriers(n int) []*barrier {
	barriers := []*barrier{}

	for i := 0; i < n; i++ {
		b := NewBarrier(barrierImg)
		b.SetPosition(float64(i)*gameWidth/float64(n-1), gameHeight-b.h/2)
		barriers = append(barriers, b)
	}

	return barriers
}
