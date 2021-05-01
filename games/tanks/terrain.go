package tanks

import (
	"math/rand"
	"time"

	rl "github.com/MattSwanson/raylib-go/raylib"
	"github.com/ojrac/opensimplex-go"
)

const (
	maxTerrainHeight = 1000 // measured from the bottom of the screen in pixels
	smoothness       = 4    // lower = smoover - 4 is good balance
)

func generateTerrain(screenWidth, screenHeight int) (rl.Texture2D, []float64) {
	rand.Seed(time.Now().UnixNano())
	noise := opensimplex.NewNormalized(rand.Int63())
	w, h := screenWidth, 1
	heightmap := make([]float64, w*h)
	for x := 0; x < w; x++ {
		xFloat := float64(x) / float64(w)
		heightmap[x] = noise.Eval2(xFloat*smoothness, 0)*maxTerrainHeight + float64(screenHeight) - maxTerrainHeight
	}
	imgW, imgH := screenWidth, screenHeight

	pixels := make([]byte, imgW*imgH*4)
	for x := 0; x < imgW; x++ {
		for y := int(heightmap[x] - 100); y < imgH; y++ {
			if float64(y) > heightmap[x] {
				pixels[(y*4*imgW)+(x*4)+1] = 0x33
				pixels[(y*4*imgW)+(x*4)+3] = 0xff
			}
		}
	}
	img := rl.LoadTextureFromImage(rl.NewImage(pixels, int32(imgW), int32(imgH), 1, rl.UncompressedR8g8b8a8))
	return img, heightmap
}
