package tanks

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"os"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	maxPlayers = 4
)

var playerLabelFont font.Face
var instFont font.Face
var xspawns = []int{400, 800, 1200, 2000}

type Core struct {
	tanks         []*tank
	currentTurn   int
	playersJoined int
	terrainImg    *ebiten.Image
	screenWidth   int
	screenHeight  int
	wind          float64
	projectile    *projectile
	gameStarted   bool
}

func Load(sWidth, sHeight int) *Core {
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
	tanks := []*tank{}
	// for i := 0; i < maxPlayers; i++ {
	// 	tanks[i] = NewTank(fmt.Sprintf("player %d", i+1))
	// }
	terrain := generateTerrain(sWidth, sHeight)

	// place the tanks at set x positions for now,
	// y position is based on terrain
	// check pixels in the given column until we find one which
	// is not 0x00 alpha
	return &Core{
		tanks:        tanks,
		terrainImg:   terrain,
		screenWidth:  sWidth,
		screenHeight: sHeight,
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
	fmt.Println(s)
	c.tanks[num].setAngle(-math.Atan(s))
	c.tanks[num].setPosition(xpos, y)
	fmt.Println(c.tanks[num].x, c.tanks[num].y)
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
	// xspawns := []float64{400, 800, 1200, 2000}
	for _, tank := range c.tanks {
		tank.Draw(screen)
		//ebitenutil.DrawLine(screen, xspawns[k], 0, xspawns[k], 1440, color.Black)
	}
	if c.projectile != nil {
		c.projectile.Draw(screen)
	}
	if c.gameStarted {
		s := fmt.Sprintf("%s's [%d] turn. !tanks shoot <angle> <velocity>", c.tanks[c.currentTurn].playerName, c.currentTurn)
		text.Draw(screen, s, instFont, 75, 1350, color.RGBA{0x00, 0xFF, 0x00, 0xFF})
	}
	// draw a wind indicator
}

func (c *Core) Update(delta float64) error {
	if c.projectile == nil {
		return nil
	}
	// tank collision
	for k, tank := range c.tanks {
		dx := c.projectile.x - tank.x - tank.w/2.0
		dy := c.projectile.y - tank.y - tank.h/2.0
		m := math.Sqrt(dx*dx + dy*dy)
		if m <= c.projectile.radius+tank.w/2.0 {
			c.tanks = removeTank(c.tanks, k)
			c.projectile = nil
			if len(c.tanks) == 1 {
				// win screen
				// winrar(c.tanks[0])
				c.gameStarted = false
				return nil
			}
			c.advanceTurn(k)
			return nil
		}
	}

	// ground collision
	for i := 0; i < 6; i++ {
		// projectile.x + radius * cos(i * 2pi/6)
		// projectile.y + radius * sin(i * 2pi/6)
		cpx := int(c.projectile.x + radius*math.Cos(float64(i)*2.0/6.0*math.Pi))
		cpy := int(c.projectile.y + radius*math.Sin(float64(i)*2.0/6.0*math.Pi))
		if _, _, _, a := c.terrainImg.At(cpx, cpy).RGBA(); a > 0 {
			// we hit the ground
			c.projectile = nil
			c.advanceTurn(-1)
			return nil
		}
	}
	//check for oob
	if c.projectile.x < -100 || c.projectile.x > float64(c.screenWidth)+100 || c.projectile.y > float64(c.screenHeight) || c.projectile.y < -2000 {
		c.projectile = nil
		c.advanceTurn(-1)
		return nil
	}
	c.projectile.Update(delta)

	return nil
}

func (c *Core) Reset() {
	c.gameStarted = false
	c.terrainImg = generateTerrain(c.screenWidth, c.screenHeight)
	c.tanks = make([]*tank, maxPlayers)
	c.playersJoined = 0
	c.currentTurn = 0
}

func (c *Core) Shoot(player string, angle float64, totalVelocity float64) {
	if !c.gameStarted || player != c.tanks[c.currentTurn].playerName {
		return
	}
	angle = angle * math.Pi / 180.0
	pSpawnOffsetX := math.Cos(angle) * c.tanks[c.currentTurn].projectileOffsetDistance
	pSpawnOffsetY := math.Sin(angle) * c.tanks[c.currentTurn].projectileOffsetDistance
	p := NewProjectile(c.tanks[c.currentTurn].x+c.tanks[c.currentTurn].w/2+pSpawnOffsetX,
		c.tanks[c.currentTurn].y+c.tanks[c.currentTurn].h/2-pSpawnOffsetY,
		c.wind)
	vx := math.Cos(angle) * totalVelocity
	vy := -math.Sin(angle) * totalVelocity
	p.SetVelocity(vx, vy)
	c.projectile = p
}

func (c *Core) AddPlayer(playerName string) {
	if c.playersJoined == maxPlayers {
		return
	}
	t := NewTank(playerName)
	c.tanks = append(c.tanks, t)
	c.PlaceTank(c.playersJoined)
	c.playersJoined++
	if c.playersJoined == maxPlayers {
		c.gameStarted = true
	}
}

func removeTank(tanks []*tank, i int) []*tank {
	// tanks[len(tanks)-1], tanks[i] = tanks[i], tanks[len(tanks)-1]
	// return tanks[:len(tanks)-1]
	return append(tanks[:i], tanks[i+1:]...)
}
