package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

var marqueeFont rl.Font
var xlMarqueeFont rl.Font
var emoteCache map[string]*imageInfo

const (
	textSize   = 120
	xlTextSize = 512
	xlYOffset  = screenHeight / 2
	regYOffset = -5
)

type marqueeMsg struct {
	RawMessage string `json:"rawMessage"`
	Emotes     string `json:"emotes"`
}

type imageInfo struct {
	img          rl.Texture2D
	animated     bool
	frameCount   int
	delay        []int
	currentFrame int
	frameCounter int
}

type emoteIndex struct {
	start   int
	end     int
	imgInfo *imageInfo
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
	font       *rl.Font
	textSize   float32
	totalWidth int
	color      color.RGBA
	oneShot    bool
	//emotes      map[string]emoteInfo
	//image    *ebiten.Image
	sequence sequence
	xOffsets []float64
}

type sequence []interface{}

func init() {
	emoteCache = make(map[string]*imageInfo)
}

func LoadMarqueeFonts() {
	marqueeFont = rl.LoadFontEx("caskaydia.TTF", textSize, nil, 0)
	xlMarqueeFont = rl.LoadFontEx("caskaydia.TTF", xlTextSize, nil, 0)
}

func NewMarquee(speed float64, color color.RGBA, oneShot bool) *Marquee {
	var currentFont *rl.Font
	currentTextSize := float32(textSize)
	if rand.Intn(100) < 10 {
		currentFont = &xlMarqueeFont
		currentTextSize = xlTextSize
	} else {
		currentFont = &marqueeFont
	}
	return &Marquee{
		speed:    speed,
		color:    color,
		oneShot:  oneShot,
		font:     currentFont,
		textSize: currentTextSize,
	}
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
		prefixLen := 13
		if m.oneShot {
			prefixLen = 14
		}

		emoteData := strings.Split(msg.Emotes, "/")
		for _, e := range emoteData {
			split := strings.Split(e, ":")
			imgInfo, err := getImageFromCDN(split[0])
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
					start:   start - prefixLen,
					end:     end - prefixLen,
					imgInfo: imgInfo,
				})
			}
		}
		sort.Sort(eIndices)
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
			m.totalWidth += int(rl.MeasureTextEx(*m.font, txt, m.textSize, 0).X)
			offsetPoints = append(offsetPoints, float64(m.totalWidth))
			m.sequence = append(m.sequence, v.imgInfo)
			m.totalWidth += int(v.imgInfo.img.Width / int32(v.imgInfo.frameCount))
			offsetPoints = append(offsetPoints, float64(m.totalWidth))
		}
		m.sequence = append(m.sequence, strippedMsg)
		m.totalWidth += int(rl.MeasureTextEx(*m.font, strippedMsg, m.textSize, 0).X)
		m.xOffsets = offsetPoints
	} else {
		m.xOffsets = []float64{0}
		m.sequence = append(m.sequence, msg.RawMessage)
		m.totalWidth = int(rl.MeasureTextEx(*m.font, msg.RawMessage, m.textSize, 0).X)
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

func UpdateEmoteCache(delta float64) {
	for _, e := range emoteCache {
		e.Update(delta)
	}
}

func (i *imageInfo) Update(delta float64) {
	if !i.animated {
		return
	}
	i.frameCounter += int(delta)
	if i.frameCounter >= i.delay[i.currentFrame]*10 {
		i.currentFrame = (i.currentFrame + 1) % i.frameCount
		i.frameCounter = 0
	}

}

func (m *Marquee) Draw() {
	if m.on {
		for k, v := range m.sequence {
			switch thing := v.(type) {
			case string:
				rl.DrawTextEx(*m.font, thing, rl.Vector2{X: float32(m.x + m.xOffsets[k]), Y: float32(m.y)}, m.textSize, 0, rl.Color(m.color))
			case *imageInfo:
				drawX := int32(m.x + m.xOffsets[k])
				drawY := int32(m.y)
				if !thing.animated {
					rl.DrawTexture(thing.img, drawX, drawY, rl.White)
				} else {
					r := rl.Rectangle{
						X:      float32(thing.currentFrame) * float32(thing.img.Width) / float32(thing.frameCount),
						Y:      0,
						Width:  float32(thing.img.Width) / float32(thing.frameCount),
						Height: float32(thing.img.Height),
					}
					rl.DrawTextureRec(thing.img, r, rl.Vector2{X: float32(drawX), Y: float32(drawY)}, rl.White)
				}
			}
		}
	}
}

func (m *Marquee) SetSpeed(speed float64) {
	m.speed = speed
}

func getImageFromCDN(id string) (*imageInfo, error) {
	// check cache first
	if img, ok := emoteCache[id]; ok {
		return img, nil
	}

	url := fmt.Sprintf("https://static-cdn.jtvnw.net/emoticons/v2/%s/default/dark/3.0", id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var img image.Image
	var rimg *rl.Image
	info := imageInfo{frameCount: 1}
	imgType := resp.Header.Get("content-type")
	if imgType == "" {
		imgType = "image/png"
	}
	switch imgType {
	case "image/gif":
		gif, err := gif.DecodeAll(resp.Body)
		if err != nil {
			return nil, err
		}

		pixels := make([]byte, gif.Config.Height*gif.Config.Width*len(gif.Image)*4)
		for row := 0; row < gif.Config.Height; row++ {
			for inum := 0; inum < len(gif.Image); inum++ {
				for x := 0; x < gif.Config.Width; x++ {
					pixelIndex := (row * len(gif.Image) * gif.Config.Width * 4) + (inum * gif.Config.Width * 4) + (x * 4)
					pc := gif.Image[inum].At(x, row)
					r, g, b, a := pc.RGBA()
					pixels[pixelIndex] = byte(r)
					pixels[pixelIndex+1] = byte(g)
					pixels[pixelIndex+2] = byte(b)
					pixels[pixelIndex+3] = byte(a)
				}

			}
		}
		rimg = rl.NewImage(pixels, int32(gif.Config.Width*len(gif.Image)), int32(gif.Config.Height), 1, rl.UncompressedR8g8b8a8)
		info.animated = true
		info.delay = gif.Delay
		noDelay := true
		for _, v := range gif.Delay {
			if v != 0 {
				noDelay = false
				break
			}
		}
		if noDelay {
			for i := 0; i < len(info.delay); i++ {
				info.delay[i] = 7
			}
		}
		info.frameCount = len(gif.Image)
	case "image/png":
		img, _, err = image.Decode(resp.Body)
		if err != nil {
			return nil, err
		}
		rimg = rl.NewImageFromImage(img)
	}

	info.img = rl.LoadTextureFromImage(rimg)

	emoteCache[id] = &info
	return emoteCache[id], nil
}
