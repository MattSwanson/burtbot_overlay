package visuals

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/MattSwanson/burtbot_overlay/sound"
	rl "github.com/MattSwanson/raylib-go/raylib"
)

type drop struct {
	drawColor rl.Color
	bounds    rl.Vector2
	name      string
}

type dropInfoMsg struct {
	User  string
	Drops []struct {
		Quality string
		Name    string
	}
}

const dropTextSize = 64
const dropPosX = 75
const dropPosY = 75

var dropFont rl.Font
var showingDrops bool
var currentDrops []drop
var cancelTimeout context.CancelFunc

var textColors = map[string]rl.Color{
	"Unique": {R: 204, G: 204, B: 102, A: 255},
	"Set":    {R: 0, G: 255, B: 0, A: 255},
	"Magic":  {R: 102, G: 102, B: 255, A: 255},
	"Rune":   {R: 255, G: 153, B: 0, A: 255},
	"Rare":   {R: 255, G: 255, B: 102, A: 255},
	"eth":    {R: 102, G: 102, B: 102, A: 255},
}

func LoadDropsAssets() {
	dropFont = rl.LoadFontEx("./visuals/Exocet2.ttf", dropTextSize, nil)
}

func ShowDrops(j string) {
	showingDrops = false
	if cancelTimeout != nil {
		cancelTimeout()
	}
	dropsMsg := dropInfoMsg{}
	json.Unmarshal([]byte(j), &dropsMsg)
	// load the new drops in
	currentDrops = []drop{}
	for _, dropStr := range dropsMsg.Drops {
		color, ok := textColors[dropStr.Quality]
		if !ok {
			color = rl.White
		}
		d := drop{
			drawColor: color,
			bounds:    rl.MeasureTextEx(dropFont, dropStr.Name, textSize, 0),
			name:      dropStr.Name,
		}
		lower := strings.ToLower(dropStr.Name)
		if strings.Contains(lower, "gold") {
			sound.Play("gold")
		}
		if strings.Contains(lower, "rune") {
			sound.Play("rune")
		}
		if strings.Contains(lower, "scroll") {
			sound.Play("scroll")
		}
		if strings.Contains(lower, "skull") {
			sound.Play("skull")
		}
		if strings.Contains(lower, "topaz") || strings.Contains(lower, "sapphire") || strings.Contains(lower, "amethyst") ||
			strings.Contains(lower, "ruby") || strings.Contains(lower, "emerald") || strings.Contains(lower, "diamond") {
			sound.Play("gem")
		}
		currentDrops = append(currentDrops, d)
	}
	showingDrops = true
	go func() {
		time.Sleep(time.Second * 5)
		showingDrops = false
	}()
}

func DrawDrops() {
	if !showingDrops {
		return
	}
	drawY := dropPosY
	for _, drop := range currentDrops {
		drawPos := rl.Vector2{X: dropPosX, Y: float32(drawY)}
		rl.DrawRectangle(int32(drawPos.X), int32(drawPos.Y), int32(drop.bounds.X), int32(drop.bounds.Y), rl.Color{0, 0, 0, 200})
		rl.DrawTextEx(dropFont, drop.name, drawPos, textSize, 0, drop.drawColor)
		drawY += int(drop.bounds.Y)
	}
}
