package tanks

import (
	"image"
	"math"
	"net/http"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

const tankSize = 48.0

var imgCache map[string]rl.Texture2D = make(map[string]rl.Texture2D)
var refAngles = []float64{
	0,
	-math.Pi / 6,
	-math.Pi / 3,
	-math.Pi / 2,
	-2 * math.Pi / 3,
	-5 * math.Pi / 6,
	-math.Pi,
}

type tank struct {
	playerName               string
	x                        float64
	y                        float64
	cx                       float64
	cy                       float64
	w                        float64
	h                        float64
	a                        float64
	scale                    float64
	projectileOffsetDistance float64
	bounds                   bounds
	img                      rl.Texture2D
	lastShotAngle            float64
}

type bounds []edge

type edge struct {
	x0 float64
	y0 float64
	x1 float64
	y1 float64
}

func (e edge) IsLeft(x, y float64) int {
	i := int((e.x1-e.x0)*(y-e.y0) -
		(x-e.x0)*(e.y1-e.y0))
	return i
}

func NewTank(playerName string, imgURL string) *tank {
	scale := 1.0
	var img rl.Texture2D
	if cached, ok := imgCache[playerName]; ok {
		img = cached
	} else {
		resp, err := http.Get(imgURL)
		if err != nil {
			img = rl.LoadTextureFromImage(rl.GenImageColor(tankSize, tankSize, rl.Blue))
		} else {
			raw, _, _ := image.Decode(resp.Body)
			img = rl.LoadTextureFromImage(rl.NewImageFromImage(raw))
			imgCache[playerName] = img
		}
	}

	scale = tankSize / float64(img.Width)

	return &tank{
		playerName:               playerName,
		img:                      img,
		w:                        scale * float64(img.Width),
		h:                        scale * float64(img.Height),
		projectileOffsetDistance: 50,
		scale:                    scale,
	}
}

func (t *tank) setPosition(x, y float64) {
	t.x, t.y = x, y

	// Get the center of the tank based on rotation
	// which is based of the bottom middle... ?
	t.cx = t.x - math.Sin(t.a)*-t.h/2
	t.cy = t.y + math.Cos(t.a)*-t.h/2

	t.setBounds()
}

func (t *tank) setAngle(theta float64) {
	t.a = theta
	t.lastShotAngle = t.a
}

func (t *tank) setBounds() {
	bounds := make([]edge, 4)

	// Calculate the corners of the tank based on the current rotation
	p0x := t.cx - math.Cos(t.a)*-t.w/2 - math.Sin(t.a)*-t.h/2
	p0y := t.cy - math.Sin(t.a)*-t.w/2 + math.Cos(t.a)*-t.h/2

	p1x := t.cx - math.Cos(t.a)*-t.w/2 - math.Sin(t.a)*t.h/2
	p1y := t.cy - math.Sin(t.a)*-t.w/2 + math.Cos(t.a)*t.h/2

	p2x := t.cx - math.Cos(t.a)*t.w/2 - math.Sin(t.a)*t.h/2
	p2y := t.cy - math.Sin(t.a)*t.w/2 + math.Cos(t.a)*t.h/2

	p3x := t.cx - math.Cos(t.a)*t.w/2 - math.Sin(t.a)*-t.h/2
	p3y := t.cy - math.Sin(t.a)*t.w/2 + math.Cos(t.a)*-t.h/2

	bounds[0] = edge{p0x, p0y, p1x, p1y}
	bounds[1] = edge{p1x, p1y, p2x, p2y}
	bounds[2] = edge{p2x, p2y, p3x, p3y}
	bounds[3] = edge{p3x, p3y, p0x, p0y}
	t.bounds = bounds

}

func (t *tank) Draw(myTurn bool) {

	// Account for rotation of the tank
	xOffset := math.Cos(t.a)*-t.w/2 - math.Sin(t.a)*-t.h
	yOffset := math.Sin(t.a)*-t.w/2 + math.Cos(t.a)*-t.h

	textColor := rl.Red
	if myTurn {
		textColor = rl.Green
		for _, ra := range refAngles {
			rl.DrawLine(int32(t.cx), int32(t.cy), int32(t.cx+50*math.Cos(t.a+ra)-math.Sin(t.a+ra)), int32(t.cy+50*math.Sin(t.a+ra)+math.Cos(t.a+ra)), rl.Green)
		}
	}

	rl.DrawTextureEx(t.img, rl.Vector2{X: float32(t.x + xOffset), Y: float32(t.y + yOffset)}, float32(t.a*180/math.Pi), float32(t.scale), rl.White)
	rl.DrawText(t.playerName, int32(t.x+t.w/2+10), int32(t.y-t.h), 24, textColor)
}


func (t *tank) DrawTurn(p int32) {
	rl.DrawTextureEx(t.img, rl.Vector2{X: 0, Y: float32(p*tankSize)}, 0, float32(t.scale), rl.White)
	rl.DrawText(t.playerName, int32(0+t.w/2+10), p*tankSize, 24, rl.Red)
}
