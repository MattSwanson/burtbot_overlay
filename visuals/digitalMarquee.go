package visuals

import (
	rl "github.com/MattSwanson/raylib-go/raylib"
)

type dmObject struct {
	// the shape of the object
	pixels [][]bool
	// position of top left
	px int
	py int
}

const (
	dmWidth        = 106  // number of columns of pixels
	dmHeight       = 20   // number of rows of pixels
	dmPixelSize    = 24   // square pixels, so one size
	dmPixelPadding = 2    // Distance between pixels
	dmTop          = 1332 // y position of top left corner
	dmLeft         = 12   // x position of top left corner
	dmScrollSpeed  = 2    // horizontal scroll speed pixels/sec
	dmRefreshRate  = 67   // ms between updates
)

var (
	dmTimeSinceLastUpdate float64 = 0.0
	//TODO: Change pixels to floats for brightness settings
	dmPixels     = [dmHeight][dmWidth]bool{}
	dmPixelColor = rl.Red
	dmTest       = dmObject{
		px: 106,
		pixels: [][]bool{
			{true, false, true},
			{true, false, true},
			{true, true, true},
			{true, false, true},
			{true, false, true},
		},
	}
	dmTest2 = dmObject{
		px: 110,
		pixels: [][]bool{
			{true, true, true},
			{true, false, false},
			{true, true, true},
			{true, false, false},
			{true, true, true},
		},
	}
	dmTest3 = dmObject{
		px: 114,
		pixels: [][]bool{
			{true, false, false},
			{true, false, false},
			{true, false, false},
			{true, false, false},
			{true, true, true},
		},
	}
	dmTest4 = dmObject{
		px: 118,
		pixels: [][]bool{
			{true, false, false},
			{true, false, false},
			{true, false, false},
			{true, false, false},
			{true, true, true},
		},
	}
	dmTest5 = dmObject{
		px: 122,
		pixels: [][]bool{
			{true, true, true},
			{true, false, true},
			{true, false, true},
			{true, false, true},
			{true, true, true},
		},
	}

	dmObjects = []dmObject{dmTest, dmTest2, dmTest3, dmTest4, dmTest5}
)

func init() {
	//dmPixels[4][5] = true
}

func DrawDMarquee() {
	for r := 0; r < dmHeight; r++ {
		for c := 0; c < dmWidth; c++ {
			if dmPixels[r][c] {
				drawX := int32(c*dmPixelSize + dmLeft)
				drawY := int32(r*dmPixelSize + dmTop)
				//rl.DrawRectangle(drawX, drawY, dmPixelSize, dmPixelSize, dmPixelColor)
				rl.DrawCircle(drawX, drawY, dmPixelSize/2, dmPixelColor)
			}
		}
	}
}

func UpdateDMarquee(delta float64) {
	dmTimeSinceLastUpdate += delta
	if dmTimeSinceLastUpdate < dmRefreshRate {
		return
	}
	dmTimeSinceLastUpdate = 0
	for r := range dmPixels {
		for c := range dmPixels[r] {
			dmPixels[r][c] = false
		}
	}
	for i := 0; i < len(dmObjects); i++ {
		dmObjects[i].px--
		if dmObjects[i].px < 0 {
			dmObjects[i].px = dmWidth - 1
		}
	}
	// go through each object
	// "or" it's pixels with everything else to composite the img
	// this maybe useless since no objects should overlap?
	for _, obj := range dmObjects {
		for r, row := range obj.pixels {
			for c, px := range row {
				cPx := obj.px + c
				cPy := obj.py + r
				// ignore pixels out of bounds
				if cPx > dmWidth-1 || cPx < 0 {
					continue
				}
				if cPy > dmHeight-1 || cPy < 0 {
					continue
				}
				dmPixels[obj.py+r][obj.px+c] = px || dmPixels[obj.py+r][obj.px+c]
			}
		}
	}
}
