package tanks

import (
	"math/rand"
	"time"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/ojrac/opensimplex-go"
)

const (
	maxTerrainHeight = 1000 // measured from the bottom of the screen in pixels
	smoothness       = 4    // lower = smoover - 4 is good balance
)

func generateTerrain(screenWidth, screenHeight int) *ebiten.Image {
	rand.Seed(time.Now().UnixNano())
	noise := opensimplex.NewNormalized(rand.Int63())
	w, h := screenWidth, 1
	heightmap := make([]float64, w*h)
	for x := 0; x < w; x++ {
		xFloat := float64(x) / float64(w)
		heightmap[x] = noise.Eval2(xFloat*smoothness, 0)*maxTerrainHeight + float64(screenHeight) - maxTerrainHeight
	}
	imgW, imgH := screenWidth, screenHeight
	img := ebiten.NewImage(imgW, imgH)
	pixels := make([]byte, imgW*imgH*4)
	for x := 0; x < imgW; x++ {
		for y := int(heightmap[x]); y < imgH; y++ {
			pixels[(y*4*imgW)+(x*4)+1] = 0x33
			pixels[(y*4*imgW)+(x*4)+3] = 0xff
		}
	}
	img.ReplacePixels(pixels)
	return img
}
