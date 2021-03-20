package main

// gophers from https://github.com/ashleymcnamara/gophers
// using fork of github.com/hajimehoshi/ebiten/v2

import (
	"bufio"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MattSwanson/burtbot_overlay/games"
	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/audio"
	"github.com/MattSwanson/ebiten/v2/audio/wav"
	"github.com/MattSwanson/ebiten/v2/ebitenutil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//var startTime time.Time
var ga Game
var myFont font.Face

const (
	audioSampleRate = 44100
)

func init() {
	var err error
	//audio init
	ga.audioContext = audio.NewContext(44100)
	ga.showStatic = false
	ga.staticLayer = static{noiseImage: image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight))}
	ga.sounds = map[string]*audio.Player{}
	ga.sounds["eep"], err = initSound(ga.audioContext, "sounds/wildeep.wav")
	if err != nil {
		log.Fatal(err)
	}
	ga.sounds["whit"], err = initSound(ga.audioContext, "sounds/Whit.wav")
	if err != nil {
		log.Fatal(err)
	}
	ga.sounds["boing"], err = initSound(ga.audioContext, "sounds/Boing.wav")
	if err != nil {
		log.Fatal(err)
	}
	ga.sounds["quack"], err = initSound(ga.audioContext, "sounds/Quack.wav")
	if err != nil {
		log.Fatal(err)
	}
	ga.sounds["zap"], err = initSound(ga.audioContext, "sounds/Voltage.wav")
	if err != nil {
		log.Fatal(err)
	}
	ga.sounds["logjam"], err = initSound(ga.audioContext, "sounds/Logjam.wav")
	if err != nil {
		log.Fatal(err)
	}
	ga.sounds["bip"], err = initSound(ga.audioContext, "sounds/Bip.wav")
	if err != nil {
		log.Fatal(err)
	}

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
	myFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    36,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	ga.marquee = NewMarquee(15)
	xs := make([]*Sprite, maxSprites)
	ga.sprites = Sprites{sprites: xs, num: 0}
	//startTime = time.Now()
}

func initSound(ctx *audio.Context, fileName string) (*audio.Player, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	ws, err := wav.DecodeWithSampleRate(audioSampleRate, file)
	if err != nil {
		return nil, err
	}
	player, err := audio.NewPlayer(ctx, ws)
	if err != nil {
		return nil, err
	}
	return player, nil
}

type Game struct {
	sprites      Sprites
	commChannel  chan cmd
	audioContext *audio.Context
	sounds       map[string]*audio.Player
	showStatic   bool
	staticLayer  static
	gameRunning  bool
	snakeGame    *Snake
	plinko       *games.Plinko
	currentInput ebiten.Key
	bigMouse     bool
	bigMouseImg  *ebiten.Image
	marquee      *Marquee
}

type cmd struct {
	command int
	arg     string
}

const (
	SpawnGopher = iota
	HideGopher
	ShowGopher
	SizeGopher
	Quack
	KillGophs
	BigMouse
	SnakeCmd
	MarqueeCmd
	TTS

	screenWidth  = 2560
	screenHeight = 1440
	maxSprites   = 100000
)

var connMessages = []string{
	"burtbot circuits activated",
	"burtboat circuits activated",
	"birdbot circus activated",
	"botbot crocus hacktivated",
	"activated circuts, burtboot",
	"botcuts burtivated accir",
	"burtboot circuit city actioned",
	"activated burtboat circumnavigation",
	"borkbonk haircut motivated",
}

func (g *Game) Update() error {

	select {
	case key := <-g.commChannel:
		switch key.command {
		case int(ebiten.KeyUp):
			g.currentInput = ebiten.KeyUp
			if g.sprites.num != 0 {
				g.sprites.sprites[0].posY -= 4.0
			}
		case int(ebiten.KeyDown):
			g.currentInput = ebiten.KeyDown
			if g.sprites.num != 0 {
				g.sprites.sprites[0].posY += 4.0
			}
		case int(ebiten.KeyLeft):
			g.currentInput = ebiten.KeyLeft
			if g.sprites.num != 0 {
				g.sprites.sprites[0].posX -= 4.0
			}
		case int(ebiten.KeyRight):
			g.currentInput = ebiten.KeyRight
			if g.sprites.num != 0 {
				g.sprites.sprites[0].posX += 4.0
			}
		case SpawnGopher:
			if num, err := strconv.Atoi(key.arg); err != nil {
				return nil
			} else {
				g.newGopher(num)
			}
		case HideGopher:
			g.hideGopher()
		case ShowGopher:
			g.showGopher()
		case SizeGopher:
			if size, err := strconv.ParseFloat(key.arg, 64); err != nil {
				return nil
			} else {
				g.setGopherSize(size)
			}
		case KillGophs:
			g.destroyGophers()
		case Quack:
			if n, err := strconv.Atoi(key.arg); err != nil {
				return nil
			} else {
				g.quack(n)
			}
		case BigMouse:
			g.bigMouse = key.arg == "true"
		case SnakeCmd:
			if key.arg == "start" && !g.gameRunning {
				g.snakeGame.reset()
				g.gameRunning = true
			} else if key.arg == "stop" {
				g.gameRunning = false
			}
		case MarqueeCmd:
			if key.arg == "off" {
				g.marquee.enable(false)
				return nil
			} else if key.arg == "embiggen" {
				g.marquee.Embiggen()
				return nil
			} else if key.arg == "smol" {
				g.marquee.Smol()
				return nil
			}
			g.marquee.setText(key.arg)
		case TTS:
			go speak(g.audioContext, key.arg)
		}
	default:
	}
	if g.gameRunning {
		g.snakeGame.Update(g.currentInput)
		g.currentInput = 0
	}
	g.plinko.Update()
	if g.showStatic {
		g.staticLayer.Update()
	}
	if g.marquee.on {
		g.marquee.Update()
	}

	for i := 0; i < g.sprites.num; i++ {
		if err := g.sprites.sprites[i].Update(); err != nil {
			return err
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	if g.bigMouse {
		cx, cy := ebiten.CursorPosition()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(cx-(g.bigMouseImg.Bounds().Dx()/2)), float64(cy-(g.bigMouseImg.Bounds().Dy()/2)))
		screen.DrawImage(g.bigMouseImg, op)
	}
	if g.gameRunning {
		g.snakeGame.Draw(screen)
	}
	for i := 0; i < g.sprites.num; i++ {
		g.sprites.sprites[i].Draw(screen)
	}
	if g.showStatic {
		g.staticLayer.Draw(screen)
	}
	g.plinko.Draw(screen)
	// text.Draw(screen, "!go spawn 100", myFont, 49, screenHeight-399, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	// text.Draw(screen, "!go spawn 100", myFont, 50, screenHeight-400, color.RGBA{0, 0xFF, 0, 0xFF})

	g.marquee.Draw(screen)

	fps := fmt.Sprintf("FPS: %.2f", ebiten.CurrentFPS())
	ebitenutil.DebugPrint(screen, fps)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("ebiten test")
	ebiten.SetScreenTransparent(true)
	ebiten.SetMousePassThru(true)
	ebiten.SetWindowFloating(true)
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetWindowDecorated(false)
	ebiten.SetInitFocused(false)
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowPosition(0, 0)

	ga.commChannel = make(chan cmd)
	game := &ga

	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	go func(c chan cmd) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err)
			}
			go handleConnection(conn, c, game.audioContext)
		}
	}(game.commChannel)
	game.plinko = games.LoadPlinko(screenWidth, screenHeight, game.sounds)
	game.plinko.ReleaseBall()
	game.snakeGame = newSnake()
	game.bigMouseImg = sprites[2]
	// _, err = getAvailableVoices()
	// if err != nil {
	// 	log.Println(err.Error())
	// }
	//fmt.Println(voices)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func handleConnection(conn net.Conn, c chan cmd, actx *audio.Context) {
	defer conn.Close()
	fmt.Println("client connected")
	msg := connMessages[rand.Intn(len(connMessages))]
	go speak(actx, msg)
	//fmt.Fprintf(conn, "Connected to overlay\n")
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		//scalego 100
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "up":
			c <- cmd{int(ebiten.KeyUp), ""}
		case "down":
			c <- cmd{int(ebiten.KeyDown), ""}
		case "left":
			c <- cmd{int(ebiten.KeyLeft), ""}
		case "right":
			c <- cmd{int(ebiten.KeyRight), ""}
		case "spawngo":
			arg := "1"
			if len(fields) > 1 {
				arg = fields[1]
			}
			c <- cmd{SpawnGopher, arg}
		case "hidego":
			c <- cmd{HideGopher, ""}
		case "showgo":
			c <- cmd{ShowGopher, ""}
		case "sizego":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{SizeGopher, fields[1]}
		case "quack":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{Quack, fields[1]}
		case "killgophs":
			c <- cmd{KillGophs, ""}
		case "bigmouse":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{BigMouse, fields[1]}
		case "snake":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{SnakeCmd, fields[1]}
		case "setmarquee":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{MarqueeCmd, strings.Join(fields[1:], " ")}
		case "tts":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{TTS, strings.Join(fields[1:], " ")}
		}
		fmt.Println(fields)
	}
}

func (g *Game) newGopher(n int) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		if g.sprites.num == maxSprites {
			return
		}
		index := rand.Int() % len(sprites)
		newGoph := NewSprite(sprites[index])
		newGoph.SetScale(0.05) // 0.05 is good size
		newGoph.SetPosition(float64(rand.Intn(screenWidth)), float64(rand.Intn(screenHeight)))

		g.sprites.sprites[g.sprites.num] = &newGoph
		g.sprites.num++

	}
	g.sounds["eep"].Rewind()
	g.sounds["eep"].Play()
}

func (g *Game) destroyGophers() {
	if g.sprites.num > 0 {
		g.sounds["logjam"].Rewind()
		g.sounds["logjam"].Play()
	}
	g.sprites.num = 0
	g.sprites.sprites = make([]*Sprite, maxSprites)
}

func (g *Game) hideGopher() {
	if g.sprites.num == 0 {
		return
	}
	if g.sprites.sprites[0].draw {
		g.sounds["whit"].Rewind()
		g.sounds["whit"].Play()
	}
	g.sprites.sprites[0].draw = false
}

func (g *Game) showGopher() {
	if g.sprites.num == 0 {
		return
	}
	if !g.sprites.sprites[0].draw {
		g.sounds["eep"].Rewind()
		g.sounds["eep"].Play()
	}
	g.sprites.sprites[0].draw = true
}

func (g *Game) setGopherSize(size float64) {
	if g.sprites.num == 0 {
		return
	}
	if size >= g.sprites.sprites[0].objScale*2 && g.sprites.sprites[0].draw {
		g.sounds["boing"].Rewind()
		g.sounds["boing"].Play()
	}
	g.sprites.sprites[0].SetScale(size)
}

func (g *Game) quack(n int) {
	go func() {
		for i := 1; i <= n; i++ {
			g.sounds["quack"].Rewind()
			g.sounds["quack"].Play()
			time.Sleep(time.Millisecond * 200)
		}
	}()
}
