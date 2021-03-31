package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var marqueeFont font.Face
var marqueeFontXl font.Face

const (
	xlYOffset  = screenHeight / 2
	regYOffset = -5
)

type marqueeMsg struct {
	RawMessage string `json:"rawMessage"`
	Emotes     string `json:"emotes"`
}

type emoteInfo struct {
	indices []emoteIndex
	img     *ebiten.Image
}
type emoteIndex struct {
	start int
	end   int
}

type Marquee struct {
	on          bool
	speed       float64
	x           float64
	y           int
	yOffset     int
	text        string
	textBounds  image.Rectangle
	color       color.RGBA
	currentFont *font.Face
	oneShot     bool
	emotes      map[string]emoteInfo
}

func init() {
	// font init
	bs, err := os.ReadFile("caskaydia.TTF")
	if err != nil {
		log.Fatal(err)
	}
	tt, err := opentype.Parse(bs)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	marqueeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    96,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	const xldpi = 144
	marqueeFontXl, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    512,
		DPI:     xldpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func NewMarquee(speed float64, color color.RGBA, oneShot bool) *Marquee {
	var currentFont *font.Face
	if rand.Intn(100) < 10 {
		currentFont = &marqueeFontXl
	} else {
		currentFont = &marqueeFont
	}
	return &Marquee{speed: speed, currentFont: currentFont, color: color, oneShot: oneShot}
}

func (m *Marquee) enable(b bool) {
	m.on = b
}

// setText takes a json string that decodes to a marqueeMsg struct
func (m *Marquee) setText(j string) {
	msg := marqueeMsg{}
	err := json.Unmarshal([]byte(j), &msg)
	if err != nil {
		log.Println(err.Error())
		return
	}
	msgEmotes := map[string]emoteInfo{}
	if msg.Emotes != "" {
		//	2313213:14-30,42-58/
		const prefixLen = 13
		emoteData := strings.Split(msg.Emotes, "/")
		//emotes := map[string][]emoteIndex{}
		for _, e := range emoteData {
			split := strings.Split(e, ":")
			indices := strings.Split(split[1], ",")
			eIndices := make([]emoteIndex, len(indices))
			for k, i := range indices {
				nums := strings.Split(i, "-")
				start, err := strconv.Atoi(nums[0])
				if err != nil {
					log.Println(err.Error())
				}
				end, err := strconv.Atoi(nums[1])
				if err != nil {
					log.Println(err.Error())
				}
				eIndices[k] = emoteIndex{start, end}
			}
			img, err := getImageFromCDN(split[0])
			if err != nil {
				log.Fatal(err)
			}
			msgEmotes[split[0]] = emoteInfo{eIndices, img}
			fmt.Println(msgEmotes)
		}
		m.text = msg.RawMessage
	} else {
		m.text = msg.RawMessage
	}
	m.emotes = msgEmotes
	m.textBounds = text.BoundString(*m.currentFont, m.text)
	// 0 to screenHeight - m.textBounds.Dy() + m.yOffset
	m.y = rand.Intn(screenHeight-m.textBounds.Dy()) + m.textBounds.Dy()
	m.color = color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 0xff}
	m.x = screenWidth
	m.on = true
}

func (m *Marquee) Update(delta float64) error {
	m.x -= m.speed * delta / 1000.0
	if m.x+float64(m.textBounds.Dx()) < 0 {
		if m.oneShot {
			return errors.New("i'm done")
		} else {
			m.x = screenWidth
		}
	}
	return nil
}

func (m *Marquee) Draw(screen *ebiten.Image) {
	if m.on {
		//text.Draw(screen, m.text, *m.currentFont, m.x+1, m.y+1, m.color)
		text.Draw(screen, m.text, *m.currentFont, int(m.x), m.y, m.color)
	}
	if len(m.emotes) > 0 {
		for _, em := range m.emotes {
			//em.img.DrawImage(screen, nil)
			screen.DrawImage(em.img, nil)
		}
	}
}

func (m *Marquee) Embiggen() {
	m.on = false
	m.currentFont = &marqueeFontXl
	m.yOffset = xlYOffset
	m.setText(m.text)
}

func (m *Marquee) Smol() {
	m.on = false
	m.currentFont = &marqueeFont
	m.yOffset = regYOffset
	m.setText(m.text)
}

func (m *Marquee) SetSpeed(speed float64) {
	m.speed = speed
}

func getImageFromCDN(id string) (*ebiten.Image, error) {
	url := fmt.Sprintf("http://static-cdn.jtvnw.net/emoticons/v1/%s/3.0", id)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}
