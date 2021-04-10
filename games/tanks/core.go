package tanks

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"time"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/audio"
	"github.com/MattSwanson/ebiten/v2/ebitenutil"
	"github.com/MattSwanson/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var playerLabelFont font.Face
var instFont font.Face
var winnerFont font.Face
var xspawns = []int{}
var boomImg *ebiten.Image

type Core struct {
	tanks         []*tank
	currentTurn   int
	playersJoined int
	terrainImg    *ebiten.Image
	heightMap     []float64
	screenWidth   int
	screenHeight  int
	wind          float64
	projectile    *projectile
	gameStarted   bool
	gameOver      bool
	winner        string
	winnerImg     *ebiten.Image
	showBoom      bool
	boomX         float64
	boomY         float64
	boomTime      time.Time
	sounds        map[string]*audio.Player
}

func Load(sWidth, sHeight int, sounds map[string]*audio.Player) *Core {
	bs, err := os.ReadFile("caskaydia.TTF")
	if err != nil {
		log.Fatal(err)
	}
	tt, err := opentype.Parse(bs)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	playerLabelFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	instFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    48,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	winnerFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    72,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	boomImg, _, err = ebitenutil.NewImageFromFile("./images/tanks/tanks_boom.png")
	if err != nil {
		log.Fatal(err)
	}

	tanks := []*tank{}
	terrain, heightMap := generateTerrain(sWidth, sHeight)

	// place the tanks at set x positions for now,
	// y position is based on terrain
	// check pixels in the given column until we find one which
	// is not 0x00 alpha
	return &Core{
		tanks:        tanks,
		terrainImg:   terrain,
		heightMap:    heightMap,
		screenWidth:  sWidth,
		screenHeight: sHeight,
		sounds:       sounds,
	}
}

func (c *Core) PlaceTank(num int) {
	var y float64
	for j := 0; j < c.screenHeight; j++ {
		// find the first 0x00 alpha pixel
		if _, _, _, a := c.terrainImg.At(xspawns[num], j+int(c.tanks[num].w)).RGBA(); a > 0 {
			y = float64(j)
			break
		}
	}
	xpos := float64(xspawns[num]) - c.tanks[num].w/2.0
	var ymo, ypo int
	for j := 0; j < c.screenHeight; j++ {
		if _, _, _, a := c.terrainImg.At(xspawns[num]-20, j+int(c.tanks[num].w)).RGBA(); a > 0 {
			ymo = j
			break
		}
	}
	for j := 0; j < c.screenHeight; j++ {
		if _, _, _, a := c.terrainImg.At(xspawns[num]+20, j+int(c.tanks[num].w)).RGBA(); a > 0 {
			ypo = j
			break
		}
	}
	s := float64(ymo-ypo) / 40
	c.tanks[num].setAngle(math.Atan(s))
	c.tanks[num].setPosition(xpos, y)
}

func (c *Core) advanceTurn(i int) {
	if i == -1 || i > c.currentTurn {
		c.currentTurn = (c.currentTurn + 1) % len(c.tanks)
		return
	}
	if i <= c.currentTurn {
		c.currentTurn %= len(c.tanks)
	}
}

func (c *Core) Draw(screen *ebiten.Image) {
	screen.DrawImage(c.terrainImg, nil)
	for _, tank := range c.tanks {
		tank.Draw(screen)
	}
	if c.projectile != nil {
		c.projectile.Draw(screen)
	}
	if c.showBoom {
		op := ebiten.DrawImageOptions{}
		op.GeoM.Translate(c.boomX, c.boomY)
		screen.DrawImage(boomImg, &op)
	}
	if c.gameStarted {
		s := fmt.Sprintf("%s's [%d] turn. !tanks shoot <angle> <velocity>", c.tanks[c.currentTurn].playerName, c.currentTurn)
		text.Draw(screen, s, instFont, 75, 1350, color.RGBA{0x00, 0xFF, 0x00, 0xFF})
	} else {
		text.Draw(screen, "type '!tanks join' to join the game!", instFont, 75, 1350, color.RGBA{0x00, 0xFF, 0x00, 0xFF})
	}
	// draw a wind indicator
	if c.gameOver {
		s := fmt.Sprintf("%s is the winner!", c.winner)
		text.Draw(screen, s, winnerFont, 410, c.screenHeight/2, color.RGBA{0x00, 0xFF, 0x00, 0xFF})
		op := ebiten.DrawImageOptions{}
		op.GeoM.Translate(100.0, float64(c.screenHeight)/2-float64(c.winnerImg.Bounds().Dy())/2)
		screen.DrawImage(c.winnerImg, &op)
	}
}

func (c *Core) Update(delta float64) error {
	if c.projectile == nil {
		return nil
	}
	// tank collision
	targets := []int{c.currentTurn}
	for k, tank := range c.tanks {
		if c.projectile.vx > 0 && c.projectile.x-tank.cx <= 0 ||
			c.projectile.vx <= 0 && c.projectile.x-tank.cx >= 0 {
			targets = append(targets, k)
		}
	}
	for _, v := range targets {
		tank := c.tanks[v]
		// diag := math.Sqrt(tank.w*tank.w + tank.h*tank.h)
		// maxDist := diag + 2*tank.w
		// var totalDist float64
		// for _, e := range tank.bounds {
		// 	dx := e.x0 - c.projectile.x + c.projectile.radius
		// 	dy := e.y0 - c.projectile.y + c.projectile.radius
		// 	totalDist += math.Sqrt(dx*dx + dy*dy)
		// }
		// if totalDist <= maxDist {
		// 	// ded?
		// 	c.boomX, c.boomY = tank.cx-float64(boomImg.Bounds().Dx())/2, tank.cy-float64(boomImg.Bounds().Dy())/2
		// 	c.boomTime = time.Now()
		// 	c.showBoom = true
		// 	c.tanks = removeTank(c.tanks, k)
		// 	c.projectile = nil
		// 	if len(c.tanks) == 1 {
		// 		// win screen
		// 		// winrar(c.tanks[0])
		// 		c.winner = c.tanks[0].playerName
		// 		c.winnerImg = c.tanks[0].img
		// 		c.gameOver = true
		// 		c.gameStarted = false
		// 		c.sounds["indigo"].Rewind()
		// 		c.sounds["indigo"].Play()
		// 		return nil
		// 	} else {
		// 		c.sounds["sosumi"].Rewind()
		// 		c.sounds["sosumi"].Play()
		// 	}
		// 	c.advanceTurn(k)
		// 	return nil
		// }
		maxDist := math.Sqrt(tank.w*tank.w+tank.h*tank.h) + 8
		dst := math.Sqrt((tank.cx-c.projectile.x)*(tank.cx-c.projectile.x) + (tank.cy-c.projectile.y)*(tank.cy-c.projectile.y))
		// if we aren't close enough for a collision to happen don't even bother checking anymore
		if dst > maxDist {
			continue
		}

		for i := 0; i < 4; i++ {
			cpx := c.projectile.x + radius*math.Cos(float64(i)*2.0/4.0*math.Pi)
			cpy := c.projectile.y + radius*math.Sin(float64(i)*2.0/4.0*math.Pi)
			if tank.bounds[0].IsLeft(cpx, cpy) > 0 && tank.bounds[1].IsLeft(cpx, cpy) > 0 && tank.bounds[2].IsLeft(cpx, cpy) > 0 && tank.bounds[3].IsLeft(cpx, cpy) > 0 {
				c.projectile = nil
				c.boomX, c.boomY = tank.cx-float64(boomImg.Bounds().Dx())/2, tank.cy-float64(boomImg.Bounds().Dy())/2
				c.boomTime = time.Now()
				c.showBoom = true
				go func() {
					time.Sleep(time.Second)
					c.showBoom = false
				}()
				c.tanks = removeTank(c.tanks, v)
				if len(c.tanks) == 1 {
					// win screen
					c.winner = c.tanks[0].playerName
					c.winnerImg = c.tanks[0].img
					c.gameOver = true
					c.gameStarted = false
					c.sounds["indigo"].Rewind()
					c.sounds["indigo"].Play()
					return nil
				} else {
					c.sounds["sosumi"].Rewind()
					c.sounds["sosumi"].Play()
				}
				c.advanceTurn(v)
				return nil
			}
		}
	}

	//check for oob
	if c.projectile.x < -100 || c.projectile.x > float64(c.screenWidth)+100 || c.projectile.y > float64(c.screenHeight) || c.projectile.y < -2000 {
		c.projectile = nil
		c.advanceTurn(-1)
		return nil
	}

	// ground collision
	for i := 0; i < 6; i++ {
		// projectile.x + radius * cos(i * 2pi/6)
		// projectile.y + radius * sin(i * 2pi/6)
		cpx := int(c.projectile.x + radius*math.Cos(float64(i)*2.0/6.0*math.Pi))
		if cpx < 0 || cpx > 2560 {
			break
		}
		cpy := c.projectile.y + radius*math.Sin(float64(i)*2.0/6.0*math.Pi)
		if cpy >= c.heightMap[cpx] {
			c.projectile = nil
			c.advanceTurn(-1)
			return nil
		}
	}

	c.projectile.Update(delta)

	return nil
}

func (c *Core) Reset() {
	c.gameStarted = false
	c.gameOver = false
	c.terrainImg, c.heightMap = generateTerrain(c.screenWidth, c.screenHeight)
	c.tanks = []*tank{}
	c.playersJoined = 0
	c.currentTurn = 0
	c.showBoom = false
}

func (c *Core) Shoot(player string, angle float64, totalVelocity float64) {
	if !c.gameStarted || player != c.tanks[c.currentTurn].playerName {
		return
	}
	angle = angle*math.Pi/180.0 + c.tanks[c.currentTurn].a
	pSpawnOffsetX := math.Cos(angle) * c.tanks[c.currentTurn].projectileOffsetDistance
	pSpawnOffsetY := math.Sin(angle) * c.tanks[c.currentTurn].projectileOffsetDistance
	c.tanks[c.currentTurn].lastShotAngle = angle
	p := NewProjectile(c.tanks[c.currentTurn].cx+pSpawnOffsetX,
		c.tanks[c.currentTurn].cy-pSpawnOffsetY,
		c.wind, false)
	vx := math.Cos(angle) * totalVelocity
	vy := -math.Sin(angle) * totalVelocity
	p.SetVelocity(vx, vy)
	c.projectile = p
}

func (c *Core) AddPlayer(playerName string, imgURL string) {
	if c.gameStarted {
		return
	}
	t := NewTank(playerName, imgURL)
	t.setPosition(0, t.h*float64(c.playersJoined))
	c.tanks = append(c.tanks, t)
	c.playersJoined++
}

func (c *Core) Begin() {
	if c.playersJoined < 2 {
		return
	}

	// generate xpos
	xspawns = []int{300}
	if c.playersJoined > 2 {
		gap := 1960 / (c.playersJoined - 1)
		for i := 1; i < c.playersJoined-1; i++ {
			x := 300 + gap*i
			xspawns = append(xspawns, x)
		}
	}
	xspawns = append(xspawns, 2260)
	c.gameStarted = true

	for i := 0; i < c.playersJoined; i++ {
		c.PlaceTank(i)
	}
}

func removeTank(tanks []*tank, i int) []*tank {
	return append(tanks[:i], tanks[i+1:]...)
}
