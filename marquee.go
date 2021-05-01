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
	"sort"
	"strconv"
	"strings"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

var marqueeFont rl.Font
var emoteCache map[string]rl.Texture2D

const (
	textSize   = 96
	xlTextSize = 512
	xlYOffset  = screenHeight / 2
	regYOffset = -5
)

type marqueeMsg struct {
	RawMessage string `json:"rawMessage"`
	Emotes     string `json:"emotes"`
}

// type emoteInfo struct {
// 	indices []emoteIndex
// 	img     *ebiten.Image
// }

type emoteIndex struct {
	start int
	end   int
	img   rl.Texture2D
}

type emoteIndices []emoteIndex

func (e emoteIndices) Len() int           { return len(e) }
func (e emoteIndices) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e emoteIndices) Less(i, j int) bool { return e[i].start < e[j].start }

type Marquee struct {
	on         bool
	speed      float64
	x          float64
	y          float64
	text       string
	totalWidth int
	textBounds image.Rectangle
	color      color.RGBA
	oneShot    bool
	//emotes      map[string]emoteInfo
	//image    *ebiten.Image
	sequence sequence
	xOffsets []float64
}

type sequence []interface{}

func init() {
	emoteCache = make(map[string]rl.Texture2D)
}

func LoadMarqueeFonts() {
	marqueeFont = rl.LoadFont("caskaydia.TTF")
}

func NewMarquee(speed float64, color color.RGBA, oneShot bool) *Marquee {
	//var currentFont *font.Face
	// if rand.Intn(100) < 10 {
	// 	currentFont = &marqueeFontXl
	// } else {
	// 	currentFont = &marqueeFont
	// }
	return &Marquee{speed: speed, color: color, oneShot: oneShot}
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
	m.xOffsets = []float64{}
	m.sequence = sequence{}
	eIndices := emoteIndices{}
	if msg.Emotes != "" {
		const prefixLen = 13
		emoteData := strings.Split(msg.Emotes, "/")

		for _, e := range emoteData {
			split := strings.Split(e, ":")
			img, err := getImageFromCDN(split[0])
			if err != nil {
				log.Fatal(err)
			}
			indices := strings.Split(split[1], ",")
			//eIndices := make([]emoteIndex, len(indices))
			for _, i := range indices {
				nums := strings.Split(i, "-")
				start, err := strconv.Atoi(nums[0])
				if err != nil {
					log.Println(err.Error())
				}
				end, err := strconv.Atoi(nums[1])
				if err != nil {
					log.Println(err.Error())
				}
				eIndices = append(eIndices, emoteIndex{
					start: start - prefixLen,
					end:   end - prefixLen,
					img:   img,
				})
			}
		}
		sort.Sort(eIndices)
		fmt.Println(eIndices)
		var offset int
		strippedMsg := msg.RawMessage
		offsetPoints := []float64{0}
		for _, v := range eIndices {
			var txt string
			if v.start-offset > len(strippedMsg) {
				fmt.Println(strippedMsg)
			}
			txt, strippedMsg = strippedMsg[:v.start-offset], strippedMsg[v.end-offset+1:]
			offset += v.end - v.start + len(txt) + 1
			txt = strings.Trim(txt, " ")
			m.sequence = append(m.sequence, txt)
			m.totalWidth += int(rl.MeasureTextEx(marqueeFont, txt, textSize, 0).X)
			offsetPoints = append(offsetPoints, float64(m.totalWidth))
			m.sequence = append(m.sequence, v.img)
			m.totalWidth += int(v.img.Width)
			offsetPoints = append(offsetPoints, float64(m.totalWidth))
		}
		m.sequence = append(m.sequence, strippedMsg)
		m.totalWidth += int(rl.MeasureTextEx(marqueeFont, strippedMsg, textSize, 0).X)
		m.xOffsets = offsetPoints
	} else {
		m.xOffsets = []float64{0}
		m.sequence = append(m.sequence, msg.RawMessage)
		m.totalWidth = int(rl.MeasureTextEx(marqueeFont, msg.RawMessage, textSize, 0).X)
		//m.totalWidth = m.textBounds.Dx()
	}
	m.text = msg.RawMessage
	//m.emotes = msgEmotes
	// 0 to screenHeight - m.textBounds.Dy() + m.yOffset
	m.y = float64(rand.Intn(screenHeight - 100))
	m.color = color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 0xff}
	m.x = screenWidth

	m.on = true
}

func (m *Marquee) Update(delta float64) error {
	m.x -= m.speed * delta / 1000.0
	if m.x+float64(m.totalWidth) < 0 {
		if m.oneShot {
			return errors.New("i'm done")
		} else {
			m.x = screenWidth
		}
	}
	return nil
}

func (m *Marquee) Draw() {
	if m.on {
		for k, v := range m.sequence {
			switch thing := v.(type) {
			case string:
				rl.DrawTextEx(marqueeFont, thing, rl.Vector2{X: float32(m.x + m.xOffsets[k]), Y: float32(m.y)}, textSize, 0, rl.Color(m.color))
			case rl.Texture2D:
				rl.DrawTexture(thing, int32(m.x+m.xOffsets[k]+25), int32(m.y), rl.White)
			}
		}
	}
}

func (m *Marquee) SetSpeed(speed float64) {
	m.speed = speed
}

func getImageFromCDN(id string) (rl.Texture2D, error) {
	// check cache first
	if img, ok := emoteCache[id]; ok {
		return img, nil
	}

	url := fmt.Sprintf("http://static-cdn.jtvnw.net/emoticons/v1/%s/3.0", id)
	resp, err := http.Get(url)
	if err != nil {
		return rl.Texture2D{}, err
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return rl.Texture2D{}, err
	}

	emoteCache[id] = rl.LoadTextureFromImage(rl.NewImageFromImage(img))
	return emoteCache[id], nil
}
