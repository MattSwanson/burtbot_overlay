package main

// gophers from https://github.com/ashleymcnamara/gophers
// using fork of github.com/hajimehoshi/ebiten/v2

import (
	"bufio"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MattSwanson/burtbot_overlay/games/lightsout"
	"github.com/MattSwanson/burtbot_overlay/games/plinko"
	"github.com/MattSwanson/burtbot_overlay/games/tanks"
	"github.com/MattSwanson/burtbot_overlay/sound"
	"github.com/MattSwanson/burtbot_overlay/visuals"
	"golang.org/x/net/context"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

//var startTime time.Time
var ga Game
var mpos rl.Vector2
var mwhipImg rl.Texture2D
var acceptedHosts []string

const (
	listenAddr = ":8081"
)

func init() {
	xs := make([]*Sprite, maxSprites)
	ga.sprites = Sprites{sprites: xs, num: 0, screenWidth: screenWidth, screenHeight: screenHeight}
	ga.lastUpdate = time.Now()

	bs, err := os.ReadFile("./accepted_hosts")
	if err != nil {
		log.Fatalln("couldn't load accepted hosts")
	}
	s := string(bs)
	addrs := strings.Fields(s)
	for _, addr := range addrs {
		if net.ParseIP(addr) == nil {
			log.Fatalln("Invalid IP address in accepted_hosts file")
		}
	}
	acceptedHosts = addrs
}

type Game struct {
	sprites         Sprites
	commChannel     chan cmd
	connWriteChan   chan string
	showStatic      bool
	staticLayer     static
	gameRunning     bool
	snakeGame       *Snake
	plinko          *plinko.Core
	plinkoRunning   bool
	tanks           *tanks.Core
	tanksRunning    bool
	lightsout       *lightsout.Core
	lightsRunning   bool
	currentInput    int
	bigMouse        bool
	bigMouseImg     rl.Texture2D
	marquees        []*Marquee
	marqueesEnabled bool
	bopometer       *visuals.Bopometer
	bingoOverlay    *visuals.BingoOverlay
	lastUpdate      time.Time
	showWhip        bool
	errorManager    *visuals.ErrorManager
}

type cmd struct {
	command int
	args    []string
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
	SingleMarqueeCmd
	TTS
	PlinkoCmd
	TanksCmd
	BopCmd
	MiracleCmd
	LightsOutCmd
	BingoCmd
	LightsCmd
	ErrorCmd

	screenWidth  = 2560
	screenHeight = 1440
	maxSprites   = 1000
)

var connMessages = []string{
	"burtbot circuits activated",
	"burtboat circuits activated",
	"birdbot circus activated",
	"botbot crocus hacktivated",
	"activated circuts, burtboot",
	"botcuts burtivated dishwasher",
	"burtboot circuit city actioned",
	"activated burtboat circumnavigation",
	"borkbonk haircut motivated",
}

func (g *Game) Update() {
	delta := float64(time.Since(g.lastUpdate).Milliseconds())
	select {
	case key := <-g.commChannel:
		switch key.command {
		case int(rl.KeyUp):
			g.currentInput = rl.KeyUp
		case int(rl.KeyDown):
			g.currentInput = rl.KeyDown
		case int(rl.KeyLeft):
			g.currentInput = rl.KeyLeft
		case int(rl.KeyRight):
			g.currentInput = rl.KeyRight

		case SpawnGopher:
			if num, err := strconv.Atoi(key.args[0]); err != nil {
				return
			} else {
				g.newGopher(num)
			}
		case HideGopher:
			g.hideGopher()
		case ShowGopher:
			g.showGopher()
		case SizeGopher:
			if size, err := strconv.ParseFloat(key.args[0], 64); err != nil {
				return
			} else {
				g.setGopherSize(size)
			}
		case KillGophs:
			g.destroyGophers()
		case Quack:
			if n, err := strconv.Atoi(key.args[0]); err != nil {
				return
			} else {
				g.quack(n)
			}
		case BigMouse:
			g.bigMouse = key.args[0] == "true"
		case SnakeCmd:
			if key.args[0] == "start" && !g.gameRunning {
				g.snakeGame.reset()
				g.gameRunning = true
			} else if key.args[0] == "stop" {
				g.gameRunning = false
			} else if key.args[0] == "speed" && len(key.args) > 1 {
				if n, err := strconv.Atoi(key.args[1]); err == nil {
					g.snakeGame.SetGameSpeed(n)
				}
			}
		case MarqueeCmd:
			if key.args[0] == "off" {
				g.marqueesEnabled = false
				g.marquees = []*Marquee{}
				return
			}
			m := NewMarquee(float64(rand.Intn(250)+450), color.RGBA{0x00, 0xff, 0x00, 0xff}, false)
			m.setText(key.args[0])
			g.marquees = append(g.marquees, m)
			g.marqueesEnabled = true
		case SingleMarqueeCmd:
			m := NewMarquee(float64(rand.Intn(250)+450), color.RGBA{0x00, 0xff, 0x00, 0xff}, true)
			m.setText(key.args[0])
			g.marquees = append(g.marquees, m)
			g.marqueesEnabled = true
		case TTS:
			cache, _ := strconv.ParseBool(key.args[1])
			go speak(key.args[0], cache)
		case PlinkoCmd:
			// !plinko start - this will start the game
			// keep alive for 60 seconds with no drops
			// each drop sets the keepalive back to 60?
			// no other arguments used
			if g.plinkoRunning {
				// !plinko drop n username
				// drop a token at drop position n for the given username
				if key.args[0] == "drop" {
					color := "#0000FF"
					if len(key.args) < 3 {
						return
					}
					if len(key.args) >= 4 {
						color = key.args[3]
					}
					// make sure we get an integer for drop position
					n, err := strconv.Atoi(key.args[1])
					if err != nil {
						// for testing:
						if key.args[1] == "all" {
							g.plinko.DropAll(key.args[2], color)
						}
						return
					}
					g.plinko.DropBall(n, key.args[2], color)
				}
			}
		case TanksCmd:
			if key.args[0] == "start" {
				g.tanksRunning = true
			} else if key.args[0] == "stop" {
				g.tanksRunning = false
				g.tanks.Reset()
			} else if key.args[0] == "join" {
				if len(key.args) < 3 {
					return
				}
				g.tanks.AddPlayer(key.args[1], key.args[2])
			} else if key.args[0] == "reset" {
				g.tanks.Reset()
			} else if key.args[0] == "shoot" {
				a, err := strconv.ParseFloat(key.args[2], 64)
				if err != nil {
					return
				}
				v, err := strconv.ParseFloat(key.args[3], 64)
				if err != nil {
					return
				}
				g.tanks.Shoot(key.args[1], a, v)
			} else if key.args[0] == "begin" {
				g.tanks.Begin()
			}
		case BopCmd:
			if key.args[0] == "start" && !g.bopometer.IsRunning() {
				g.bopometer.Reset()
				g.bopometer.SetRunning(true)
			} else if key.args[0] == "add" && g.bopometer.IsRunning() {
				if len(key.args) < 2 {
					return
				}
				n, err := strconv.Atoi(key.args[1])
				if err != nil {
					return
				}
				g.bopometer.Add(n)
			} else if key.args[0] == "stop" && g.bopometer.IsRunning() {
				g.bopometer.Finish()
				g.bopometer.SetRunning(false)
			}
		case LightsOutCmd:
			if key.args[0] == "start" && !g.lightsRunning {
				g.lightsout.LoadPuzzle(0)
				g.lightsRunning = true
				break
			}
			if g.lightsRunning {
				if key.args[0] == "reset" {
					g.lightsout.Reset()
					break
				}
				if key.args[0] == "stop" {
					g.lightsRunning = false
					break
				}
				n, err := strconv.Atoi(key.args[0])
				if err != nil {
					break
				}
				g.lightsout.Press(n)
			}
		case BingoCmd:
			if key.args[0] == "drawn" {
				if len(key.args) < 2 {
					break
				}
				g.bingoOverlay.AddNumber(key.args[1])
			} else if key.args[0] == "reset" {
				g.bingoOverlay.Reset()
			} else if key.args[0] == "winner" {
				if len(key.args) < 3 {
					break
				}
				g.bingoOverlay.End(key.args[1], key.args[2])
			}
		case MiracleCmd:
			g.showWhip = true
			sound.Play("indigo")
			go func() {
				time.Sleep(time.Second * 5)
				g.showWhip = false
			}()
		case LightsCmd:
			if key.args[0] == "set" {
				color, err := strconv.Atoi(key.args[1])
				if err != nil {
					break
				}
				visuals.SetLightsColor(color)
			}
		case ErrorCmd:
			g.errorManager.AddError(5)
			go func() {
				time.Sleep(time.Second * 5)
				g.errorManager.Clear()
			}()
		}
	default:
	}
	if g.gameRunning {
		g.snakeGame.Update(g.currentInput)
		g.currentInput = 0
	}
	if g.plinkoRunning {
		g.plinko.Update()
	}
	if g.showStatic {
		g.staticLayer.Update()
	}
	if g.bopometer.IsRunning() {
		g.bopometer.Update(delta)
	}
	if g.marqueesEnabled {
		for i := 0; i < len(g.marquees); i++ {
			if err := g.marquees[i].Update(delta); err != nil {
				copy(g.marquees[i:], g.marquees[i+1:])
				g.marquees[len(g.marquees)-1] = nil
				g.marquees = g.marquees[:len(g.marquees)-1]
			}
		}
	}
	if g.tanksRunning {
		g.tanks.Update(delta)
	}
	if g.errorManager.Visible {
		g.errorManager.Update(delta)
	}

	for i := 0; i < g.sprites.num; i++ {
		if err := g.sprites.sprites[i].Update(delta); err != nil {
			return
		}
	}
	g.lastUpdate = time.Now()
}

func (g *Game) Draw() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.Color{R: 0x00, G: 0x00, B: 0x00, A: 0x00})
	rl.DrawFPS(50, 50)

	g.errorManager.Draw()

	if g.bigMouse {
		rl.DrawTexture(g.bigMouseImg, int32(mpos.X), int32(mpos.Y), rl.White)
	}

	if g.tanksRunning {
		g.tanks.Draw()
	}

	if g.lightsRunning {
		g.lightsout.Draw()
	}

	if g.gameRunning {
		g.snakeGame.Draw()
	}
	for i := 0; i < g.sprites.num; i++ {
		g.sprites.sprites[i].Draw()
	}
	// if g.showStatic {
	// 	g.staticLayer.Draw(screen)
	// }
	if g.plinkoRunning {
		g.plinko.Draw()
	}
	if g.bopometer.IsRunning() || g.bopometer.IsFinished() {
		g.bopometer.Draw()
	}

	if g.marqueesEnabled {
		for i := 0; i < len(g.marquees); i++ {
			g.marquees[i].Draw()
		}
	}

	if g.showWhip {
		rl.DrawTextureEx(mwhipImg, rl.Vector2{X: 560, Y: 0}, 0, 0.6, rl.White)
	}

	g.bingoOverlay.Draw()

	rl.EndDrawing()
}

func main() {
	rl.SetConfigFlags(rl.FlagWindowFloating | rl.FlagWindowMousePassthrough | rl.FlagWindowTransparent | rl.FlagWindowUndecorated)
	rl.InitWindow(screenWidth, screenHeight, "burtbot overlay")
	rl.SetTargetFPS(60)
	rl.InitAudioDevice()
	rl.SetMasterVolume(sound.MasterVolume)

	mwhipImg = rl.LoadTexture("./images/mwhip.png")
	LoadSprites()
	visuals.LoadBopometerAssets()
	ga.commChannel = make(chan cmd)
	ga.connWriteChan = make(chan string)
	game := &ga
	game.bigMouseImg = sprites[2]
	LoadMarqueeFonts()

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	go func(c chan cmd, wc chan string) {
		for {
			conn, err := ln.Accept()
			fmt.Printf("Connection from %s\n", conn.RemoteAddr().String())
			if err != nil {
				log.Println(err)
			}
			remoteAddress := strings.Split(conn.RemoteAddr().String(), ":")
			acceptedHost := false
			for _, addr := range acceptedHosts {
				if remoteAddress[0] == addr {
					acceptedHost = true
					break
				}
			}
			if !acceptedHost {
				go speak("Intrusion Detected", true)
				conn.Close()
				continue
			}
			go handleConnection(conn, c, wc)
		}
	}(game.commChannel, ga.connWriteChan)
	game.plinko = plinko.Load(screenWidth, screenHeight, game.connWriteChan)
	defer game.plinko.CancelTimer()
	game.plinkoRunning = true
	game.snakeGame = newSnake()
	game.tanks = tanks.Load(screenWidth, screenHeight)
	game.bopometer = visuals.NewBopometer(game.connWriteChan)
	game.lightsout = lightsout.NewGame(5, 5)
	game.bingoOverlay = visuals.NewBingoOverlay()
	game.errorManager = visuals.NewErrorManager()

	for !rl.WindowShouldClose() {
		game.Update()
		game.Draw()
	}
	rl.CloseAudioDevice()
	rl.CloseWindow()
}

func handleConnection(conn net.Conn, c chan cmd, wc chan string) {
	defer conn.Close()
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		handleWrites(ctx, &conn, wc)
	}(ctx)
	defer cancel()
	fmt.Println("client connected")
	msg := connMessages[rand.Intn(len(connMessages))]
	go speak(msg, true)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Fields(txt)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "up":
			c <- cmd{int(rl.KeyUp), []string{}}
		case "down":
			c <- cmd{int(rl.KeyDown), []string{}}
		case "left":
			c <- cmd{int(rl.KeyLeft), []string{}}
		case "right":
			c <- cmd{int(rl.KeyRight), []string{}}
		case "spawngo":
			arg := "1"
			if len(fields) > 1 {
				arg = fields[1]
			}
			c <- cmd{SpawnGopher, []string{arg}}
		case "hidego":
			c <- cmd{HideGopher, []string{}}
		case "showgo":
			c <- cmd{ShowGopher, []string{}}
		case "sizego":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{SizeGopher, fields[1:]}
		case "quack":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{Quack, fields[1:]}
		case "killgophs":
			c <- cmd{KillGophs, []string{}}
		case "bigmouse":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{BigMouse, fields[1:]}
		case "snake":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{SnakeCmd, fields[1:]}
		case "marquee":
			if fields[1] == "off" {
				c <- cmd{MarqueeCmd, []string{"off"}}
			}
			if len(fields) < 3 {
				continue
			}
			if fields[1] == "set" {
				c <- cmd{MarqueeCmd, []string{txt[12:]}}
			} else if fields[1] == "once" {
				c <- cmd{SingleMarqueeCmd, []string{txt[12:]}}
			}
		case "tts":
			if len(fields) < 3 {
				continue
			}
			c <- cmd{TTS, []string{strings.Join(fields[2:], " "), fields[1]}}
		case "plinko":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{PlinkoCmd, fields[1:]}
		case "tanks":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{TanksCmd, fields[1:]}
		case "bop":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{BopCmd, fields[1:]}
		case "miracle":
			c <- cmd{MiracleCmd, []string{}}
		case "lo":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{LightsOutCmd, fields[1:]}
		case "bingo":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{BingoCmd, fields[1:]}
			// j := fmt.Sprintf(`{"RawMessage":"%s", "Emotes":""}`, fields[2])
			// c <- cmd{SingleMarqueeCmd, []string{j}}
		case "lights":
			if len(fields) < 3 {
				continue
			}
			c <- cmd{LightsCmd, fields[1:]}
		case "error":
			c <- cmd{ErrorCmd, []string{}}
		}

		fmt.Println(fields)
	}
}

func handleWrites(ctx context.Context, conn *net.Conn, wc chan string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Canceling tcp write loop")
			return
		case s := <-wc:
			n, err := fmt.Fprint(*conn, s)
			if err != nil {
				log.Println("couldn't write to connection: ", err.Error())
				break
			}
			log.Printf("wrote %d bytes to tcp connection", n)
		}
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
	sound.Play("eep")
}

func (g *Game) destroyGophers() {
	if g.sprites.num > 0 {
		sound.Play("logjam")
	}
	g.sprites.num = 0
	g.sprites.sprites = make([]*Sprite, maxSprites)
}

func (g *Game) hideGopher() {
	if g.sprites.num == 0 {
		return
	}
	if g.sprites.sprites[0].draw {
		sound.Play("whit")
	}
	g.sprites.sprites[0].draw = false
}

func (g *Game) showGopher() {
	if g.sprites.num == 0 {
		return
	}
	if !g.sprites.sprites[0].draw {
		sound.Play("eep")
	}
	g.sprites.sprites[0].draw = true
}

func (g *Game) setGopherSize(size float64) {
	if g.sprites.num == 0 {
		return
	}
	if size >= g.sprites.sprites[0].objScale*2 && g.sprites.sprites[0].draw {
		sound.Play("boing")
	}
	g.sprites.sprites[0].SetScale(size)
}

func (g *Game) quack(n int) {
	go func() {
		for i := 1; i <= n; i++ {
			sound.Play("quack")
			time.Sleep(time.Millisecond * 200)
		}
	}()
}
