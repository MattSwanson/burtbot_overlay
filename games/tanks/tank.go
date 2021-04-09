package tanks

import (
	"image"
	"image/color"
	"net/http"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/ebitenutil"
	"github.com/MattSwanson/ebiten/v2/text"
	"gonum.org/v1/gonum/mat"
)

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
	img                      *ebiten.Image
}

type bounds []edge

type edge struct {
	x0 float64
	y0 float64
	x1 float64
	y1 float64
}

func (e edge) IsLeft(x, y float64) int {
	return int((e.x1-e.x0)*(y-e.y0) - (x-e.x0)*(e.y1-e.y0))
}

func NewTank(playerName string, imgURL string) *tank {
	scale := 1.0
	img := ebiten.NewImage(48, 48)
	resp, err := http.Get(imgURL)
	if err != nil {
		img.Fill(color.RGBA{0x00, 0x00, 0xff, 0xff})
	} else {
		raw, _, err := image.Decode(resp.Body)
		if err != nil {
			img.Fill(color.RGBA{0x00, 0x00, 0xff, 0xff})
		}
		img = ebiten.NewImageFromImage(raw)
		scale = 48 / float64(img.Bounds().Dx())
	}

	return &tank{
		playerName:               playerName,
		img:                      img,
		w:                        scale * float64(img.Bounds().Dx()),
		h:                        scale * float64(img.Bounds().Dy()),
		projectileOffsetDistance: 45,
		scale:                    scale,
	}
}

func (t *tank) setPosition(x, y float64) {
	t.x, t.y = x, y
	t.setBounds()
}

func (t *tank) setAngle(theta float64) {
	t.a = theta
	t.setBounds()
}

func (t *tank) setBounds() {
	bounds := make([]edge, 4)
	pos := mat.NewDense(3, 1, []float64{t.x, t.y, 1})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-t.x-t.w/2, -t.y-t.h)
	op.GeoM.Rotate(t.a)
	op.GeoM.Translate(t.x+t.w/2, t.y+t.h)

	final := mat.NewDense(3, 3, []float64{
		op.GeoM.Element(0, 0),
		op.GeoM.Element(0, 1),
		op.GeoM.Element(0, 2),
		op.GeoM.Element(1, 0),
		op.GeoM.Element(1, 1),
		op.GeoM.Element(1, 2),
		0,
		0,
		1,
	})
	var fp1 mat.Dense
	fp1.Mul(final, pos)

	pos2 := mat.NewDense(3, 1, []float64{t.x + t.w, t.y, 1})
	var fp2 mat.Dense
	fp2.Mul(final, pos2)

	pos3 := mat.NewDense(3, 1, []float64{t.x + t.w, t.y + t.h, 1})
	var fp3 mat.Dense
	fp3.Mul(final, pos3)

	pos4 := mat.NewDense(3, 1, []float64{t.x, t.y + t.h, 1})
	var fp4 mat.Dense
	fp4.Mul(final, pos4)

	center := mat.NewDense(3, 1, []float64{t.x + t.w/2, t.y + t.h/2, 1})
	var cprime mat.Dense
	cprime.Mul(final, center)
	t.cx, t.cy = cprime.At(0, 0), cprime.At(1, 0)
	bounds[0] = edge{fp1.At(0, 0), fp1.At(1, 0), fp2.At(0, 0), fp2.At(1, 0)}
	bounds[1] = edge{fp2.At(0, 0), fp2.At(1, 0), fp3.At(0, 0), fp3.At(1, 0)}
	bounds[2] = edge{fp3.At(0, 0), fp3.At(1, 0), fp4.At(0, 0), fp4.At(1, 0)}
	bounds[3] = edge{fp4.At(0, 0), fp4.At(1, 0), fp1.At(0, 0), fp1.At(1, 0)}
	t.bounds = bounds

}

func (t *tank) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(t.img.Bounds().Dx())/2, -float64(t.img.Bounds().Dy())/2)
	op.GeoM.Scale(t.scale, t.scale)
	op.GeoM.Translate(0, -t.h/2)
	op.GeoM.Rotate(t.a)
	op.GeoM.Translate(t.w/2, t.h)
	op.GeoM.Translate(t.x, t.y)
	screen.DrawImage(t.img, op)
	text.Draw(screen, t.playerName, playerLabelFont, int(t.x+t.w+5), int(t.y+t.h/2+5), color.RGBA{0xff, 0x00, 0x00, 0xff})
	for _, e := range t.bounds {
		drawEdge(screen, e, color.RGBA{0x00, 0x00, 0xff, 0xff})
	}
}

func drawEdge(screen *ebiten.Image, e edge, c color.Color) {
	ebitenutil.DrawLine(screen, e.x0, e.y0, e.x1, e.y1, c)
}
