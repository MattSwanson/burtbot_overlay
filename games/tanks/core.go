package tanks

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/MattSwanson/burtbot_overlay/sound"
	rl "github.com/MattSwanson/raylib-go/raylib"
)

const (
	maxShotVelocity = 1500 // ?
	slopeCalcOffset = 20
)

var boomImg rl.Texture2D

type Core struct {
	tanks         []*tank
	currentTurn   int
	turnOrder     []*tank
	playersJoined int
	terrainImg    rl.Texture2D
	heightMap     []float64
	screenWidth   int
	screenHeight  int
	wind          float64
	projectile    *projectile
	gameStarted   bool
	gameOver      bool
	winner        string
	winnerImg     rl.Texture2D
	showBoom      bool
	boomX         float64
	boomY         float64
	boomTime      time.Time
	running       bool
}

func rotateTurns(s []*tank) []*tank {
	return append(s[1:], s[0])
}

func (c *Core) removeTankFromTurnOrder(t *tank) {
	tanks := []*tank{}
	for _, tank := range c.turnOrder {
		if t != tank {
			tanks = append(tanks, tank)
		}
	}
	c.turnOrder = tanks
}

func Load(sWidth, sHeight int) *Core {

	boomImg = rl.LoadTexture("./images/tanks/tanks_boom.png")

	tanks := []*tank{}
	terrain, heightMap := generateTerrain(sWidth, sHeight)

	w := (rand.Float64() - 0.5) * 100
	return &Core{
		tanks:        tanks,
		terrainImg:   terrain,
		heightMap:    heightMap,
		wind:         w,
		screenWidth:  sWidth,
		screenHeight: sHeight,
	}
}

func (c *Core) PlaceTank(num int, xpos int) {
	y := c.heightMap[xpos]
	ymo := c.heightMap[xpos-slopeCalcOffset]
	ypo := c.heightMap[xpos+slopeCalcOffset]
	s := (ymo - ypo) / (2 * slopeCalcOffset)
	c.tanks[num].setAngle(-math.Atan(s))
	c.tanks[num].setPosition(float64(xpos), y)
}

func (c *Core) Draw() {
	if !c.running {
		return
	}
	if c.gameStarted {
		for i, t := range c.turnOrder {
			t.DrawTurn(int32(i))
		}
	}
	rl.DrawTexture(c.terrainImg, 0, 0, rl.White)
	for _, tank := range c.tanks {
		myTurn := c.turnOrder[0] == tank
		tank.Draw(myTurn)
	}
	if c.projectile != nil {
		c.projectile.Draw()
	}
	if c.showBoom {
		rl.DrawTexture(boomImg, int32(c.boomX), int32(c.boomY), rl.White)
	}
	if c.gameStarted {
		s := fmt.Sprintf("%s's turn. !tanks shoot <angle(degrees)> <velocity(1-100)>", c.turnOrder[0].playerName)
		rl.DrawText(s, 75, 1350, 48, rl.Color{R: 0x00, G: 0xFF, B: 0x00, A: 0xFF})
	} else {
		rl.DrawText("type '!tanks join' to join the game!", 75, 1350, 48, rl.Color{R: 0x00, G: 0xFF, B: 0x00, A: 0xFF})
	}
	// // draw a wind indicator
	windIndX := float64(c.screenWidth) / 2
	if c.wind < 0 {
		windIndX += c.wind
	}
	rl.DrawText(fmt.Sprintf("Wind: %0.2f", c.wind), int32(c.screenWidth/2), 100, 32, rl.Green)
	rl.DrawRectangle(int32(windIndX), 0, int32(math.Abs(c.wind)), 50, rl.Blue)
	rl.DrawLine(int32(c.screenWidth/2), 0, int32(c.screenWidth/2), 75, rl.Green)
	if c.gameOver {
		s := fmt.Sprintf("%s is the winner!", c.winner)
		rl.DrawText(s, 410, int32(c.screenHeight/2), 48, rl.Color{R: 0x00, G: 0xFF, B: 0x00, A: 0xFF})
		rl.DrawTexture(c.winnerImg, 100, int32(c.screenHeight/2-int(float32(c.winnerImg.Width)/2.0)), rl.White)
	}
}

func (c *Core) HandleMessage(args []string) {
	if args[0] == "start" {
		c.running = true
	} else if args[0] == "stop" {
		c.running = false
		c.Reset()
	} else if args[0] == "join" {
		if len(args) < 3 {
			return
		}
		c.AddPlayer(args[1], args[2])
	} else if args[0] == "reset" {
		c.Reset()
	} else if args[0] == "shoot" {
		a, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return
		}
		v, err := strconv.ParseFloat(args[3], 64)
		if err != nil {
			return
		}
		c.Shoot(args[1], a, v)
	} else if args[0] == "begin" {
		c.Begin()
	}
}

func (c *Core) Update(delta float64) error {
	if !c.running {
		return nil
	}
	if c.projectile == nil {
		return nil
	}

	for v, tank := range c.tanks {
		maxDist := math.Sqrt(tank.w*tank.w+tank.h*tank.h) + 8
		dst := math.Sqrt((tank.cx-c.projectile.x)*(tank.cx-c.projectile.x) + (tank.cy-c.projectile.y)*(tank.cy-c.projectile.y))
		if dst > maxDist {
			continue
		}

		for i := 0; i < 32; i++ {
			cpx := c.projectile.x + radius*math.Cos(float64(i)*2.0/32.0*math.Pi)
			cpy := c.projectile.y - radius*math.Sin(float64(i)*2.0/32.0*math.Pi)
			if tank.bounds[0].IsLeft(cpx, cpy) > 0 && tank.bounds[1].IsLeft(cpx, cpy) > 0 && tank.bounds[2].IsLeft(cpx, cpy) > 0 && tank.bounds[3].IsLeft(cpx, cpy) > 0 {
				c.projectile = nil
				c.boomX, c.boomY = tank.cx-float64(boomImg.Width)/2, tank.cy-float64(boomImg.Width)/2
				c.boomTime = time.Now()
				c.showBoom = true
				go func() {
					time.Sleep(time.Second)
					c.showBoom = false
				}()
				c.playersJoined--
				c.tanks = removeTank(c.tanks, v)
				if len(c.tanks) == 1 {
					// win screen
					c.winner = c.tanks[0].playerName
					c.winnerImg = c.tanks[0].img
					c.gameOver = true
					c.gameStarted = false
					sound.Play("indigo")
					go func() {
						time.Sleep(5 * time.Second)
						c.gameOver = false
					}()
					c.Reset()
					return nil
				} else {
					sound.Play("sosumi")
				}

				if tank != c.turnOrder[0] {
					c.removeTankFromTurnOrder(tank)
					c.turnOrder = rotateTurns(c.turnOrder)
				} else {
					c.removeTankFromTurnOrder(tank)
				}
				return nil
			}
		}
	}

	//check for oob
	if c.projectile.x < -100 || c.projectile.x > float64(c.screenWidth)+100 || c.projectile.y > float64(c.screenHeight) || c.projectile.y < -2000 {
		c.projectile = nil
		// c.advanceTurn(-1)
		c.turnOrder = rotateTurns(c.turnOrder)
		return nil
	}

	// ground collision
	for i := 0; i < 6; i++ {
		// find the x center of the projectile from the top left corner (origin)
		cpx := int(c.projectile.x + radius*math.Cos(float64(i)*2.0/6.0*math.Pi))
		if cpx < 0 || cpx >= 2560 {
			break
		}
		// then the y center
		cpy := c.projectile.y + radius*math.Sin(float64(i)*2.0/6.0*math.Pi)
		// if the y position is lower than the height of the terrain at that x pos...
		if cpy >= c.heightMap[cpx] {
			// thunk
			sound.Play("kerplunk")
			c.projectile = nil
			//c.advanceTurn(-1)
			c.turnOrder = rotateTurns(c.turnOrder)
			return nil
		}
	}

	c.projectile.Update(delta)

	return nil
}

func (c *Core) Reset() {
	c.gameStarted = false
	c.terrainImg, c.heightMap = generateTerrain(c.screenWidth, c.screenHeight)
	c.tanks = []*tank{}
	c.playersJoined = 0
	c.turnOrder = []*tank{}
	c.wind = (rand.Float64() - 0.5) * 100
	c.showBoom = false
}

func (c *Core) Shoot(player string, angle float64, totalVelocity float64) {
	if !c.gameStarted || !strings.HasPrefix(c.turnOrder[0].playerName, player) || c.projectile != nil {
		return
	}
	if totalVelocity < 1 {
		return
	}
	totalVelocity = math.Min(totalVelocity, 100)
	totalVelocity = maxShotVelocity * totalVelocity / 100
	angle = angle*math.Pi/180.0 - c.turnOrder[0].a
	pSpawnOffsetX := math.Cos(angle) * c.turnOrder[0].projectileOffsetDistance
	pSpawnOffsetY := math.Sin(angle) * c.turnOrder[0].projectileOffsetDistance
	c.turnOrder[0].lastShotAngle = angle
	p := NewProjectile(c.turnOrder[0].cx+pSpawnOffsetX,
		c.turnOrder[0].cy-pSpawnOffsetY,
		c.wind, false)
	vx := math.Cos(angle) * totalVelocity
	vy := -math.Sin(angle) * totalVelocity
	p.SetVelocity(vx, vy)
	c.projectile = p
}

func (c *Core) AddPlayer(playerName string, imgURL string) {
	if strings.HasPrefix(playerName, "BurtStanton") {
		playerName += fmt.Sprint(len(c.tanks))
	}
	xpos := 0
	ind := 0
	t := NewTank(playerName, imgURL)
	if c.gameStarted {
		l := slopeCalcOffset - 1
		maxDist := int(c.tanks[0].x)
		r := maxDist
		for i := 0; i < c.playersJoined-1; i++ {
			d := int(c.tanks[i+1].x - c.tanks[i].x)
			if d > maxDist {
				maxDist = d
				l = int(c.tanks[i].x)
				r = int(c.tanks[i+1].x)
				ind = i + 1
			}
		}
		if c.screenWidth-slopeCalcOffset-int(c.tanks[c.playersJoined-1].x) > maxDist {
			ind = c.playersJoined
			l = int(c.tanks[c.playersJoined-1].x)
			r = c.screenWidth - slopeCalcOffset
		}
		xpos = rand.Intn(r-l) + l
		if ind == c.playersJoined {
			c.tanks = append(c.tanks, t)
		} else if ind == 0 {
			c.tanks = append([]*tank{t}, c.tanks...)
			// c.currentTurn = 1
		} else {
			tmp := make([]*tank, len(c.tanks)+1)
			for i := 0; i < len(tmp); i++ {
				if i < ind {
					tmp[i] = c.tanks[i]
				} else if i == ind {
					tmp[i] = t
				} else {
					tmp[i] = c.tanks[i-1]
				}
			}
			c.tanks = tmp
		}
	}
	c.turnOrder = append(c.turnOrder, t)
	c.playersJoined++
	if !c.gameStarted {
		t.setPosition(0+t.w/2, t.h*float64(c.playersJoined-1)+t.h)
		c.tanks = append(c.tanks, t)
		return
	}
	c.PlaceTank(ind, xpos)
}

func (c *Core) Begin() {
	if c.playersJoined < 2 {
		return
	}
	r := (c.screenWidth - 2*slopeCalcOffset) / (c.playersJoined * 2)
	for i := 0; i < c.playersJoined; i++ {
		xpos := rand.Intn(r) + r*(i*2) + slopeCalcOffset
		c.PlaceTank(i, xpos)
	}
	c.gameStarted = true
}

func removeTank(tanks []*tank, i int) []*tank {
	return append(tanks[:i], tanks[i+1:]...)
}
