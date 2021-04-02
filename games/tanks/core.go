package tanks

import (
	"fmt"
	"log"
	"os"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/inpututil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	maxPlayers = 4
)

var playerLabelFont font.Face

type Core struct {
	tanks        []*tank
	terrainImg   *ebiten.Image
	screenWidth  int
	screenHeight int
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
	tanks := make([]*tank, maxPlayers)
	for i := 0; i < maxPlayers; i++ {
		tanks[i] = NewTank(fmt.Sprintf("player %d", i+1))
	}
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

func (c *Core) PlaceTanks() {
	xspawns := []int{400, 800, 1200, 2000}
	var y float64
	for i := 0; i < 4; i++ {
		for j := 0; j < c.screenHeight; j++ {
			// find the first 0x00 alpha pixel
			if _, _, _, a := c.terrainImg.At(xspawns[i], j).RGBA(); a > 0 {
				y = float64(j)
			}
		}
		c.tanks[i].setPosition(float64(xspawns[i]), y)
		fmt.Println(c.tanks[i].x, c.tanks[i].y)
	}
}

func (c *Core) Draw(screen *ebiten.Image) {
	screen.DrawImage(c.terrainImg, nil)
	for _, t := range c.tanks {
		t.Draw(screen)
	}
}

func (c *Core) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		c.PlaceTanks()
	}
	return nil
}
