package main

// gophers from https://github.com/ashleymcnamara/gophers
// using fork of github.com/gen2brain/raylib-go/raylib

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/MattSwanson/burtbot_overlay/games"
	"github.com/MattSwanson/burtbot_overlay/games/cube"
	"github.com/MattSwanson/burtbot_overlay/planes"
	"github.com/MattSwanson/burtbot_overlay/shaders"
	"github.com/MattSwanson/burtbot_overlay/sound"
	"github.com/MattSwanson/burtbot_overlay/speech"
	"github.com/MattSwanson/burtbot_overlay/visuals"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"golang.org/x/net/context"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

// var startTime time.Time
var ga Game
var mwhipImg rl.Texture2D
var mkImg rl.Texture2D
var carImg rl.Texture2D
var acceptedHosts []string
var dedCount int

var tuxpos rl.Vector3 = rl.Vector3{X: 0, Y: 0, Z: -500}
var showtux bool
var gettingHR bool
var lastMetricsUpdate time.Time
var nowPlaying string
var npTextY float32
var npBGY int32
var goodFont rl.Font
var ibmFont rl.Font
var moos = []string{
	"moo_a1",
	"moo_a2",
	"moo_a3",
	"moo_a4",
	"moo_a5",
	"moo_a6",
	"moo_d1",
	"moo_d2",
	"moo_d3",
	"moo_d4",
	"moo_d5",
	"moo_h1",
	"moo_h2",
	"moo_h3",
	"moo_h4",
	"moo_n1",
	"moo_n2",
	"moo_n3",
	"moo_n4",
	"moo_n5",
}

// var usbDriver *ant.GarminStick3
var signalChannel chan os.Signal
var useANT = false
var obsCmd *exec.Cmd
var camera rl.Camera3D
var showPlanes = false
var isVerbose = false

const (
	listenAddr = ":8081"

	npTextTopY    = 10
	npTextBottomY = 1375
)

var goobsClient *goobs.Client
var streamHealthCancelFunc context.CancelFunc

type RTMPApplication struct {
	Name       string `xml:"name"`
	BitRate    int    `xml:"live>stream>bw_video"`
	Publishing []bool `xml:"live>stream>publishing"`
}

func init() {
	flag.BoolVar(&useANT, "a", false, "enable ANT sensor")
	flag.BoolVar(&showPlanes, "p", false, "track seen adsb planes")
	flag.BoolVar(&isVerbose, "v", false, "show all incoming messages")
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
	sprites       Sprites
	commChannel   chan cmd
	connWriteChan chan string
	showStatic    bool
	staticLayer   static
	gameRunning   bool
	snakeGame     *Snake
	currentInput  int
	bigMouse      bool
	bigMouseImg   rl.Texture2D
	bopometer     *visuals.Bopometer
	bingoOverlay  *visuals.BingoOverlay
	lastUpdate    time.Time
	showWhip      bool
	showMK        bool
	showDM        bool
	showFSInfo    bool
	errorManager  *visuals.ErrorManager
}

type cmd struct {
	command int
	args    []string
}

const (
	SpawnGopher = iota
	Quack
	KillGophs
	BigMouse
	SnakeCmd
	MarqueeCmd
	SingleMarqueeCmd
	TTS
	BopCmd
	MiracleCmd
	MKCmd
	BingoCmd
	LightsCmd
	ErrorCmd
	Quacksplosion
	FollowAlert
	DedCmd
	CubeCmd
	TuxCmd
	NowPlayingCmd
	NpTextCmd
	MooCmd
	DropsCmd
	DMCmd
	ToggleFSInfoCmd
	FSCmd
	StreamCmd
	MetricsCmd
	RaidAlert
	SteamCmd
	GameCmd

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

	if showtux {
		tuxpos.Z += float32(50.0 * delta / 1000)
		if tuxpos.Z > 25 {
			showtux = false
		}
	}
	select {
	case signal := <-signalChannel:
		if signal == os.Interrupt {
			cleanUp()
		}
	case key := <-g.commChannel:
		switch key.command {
		case int(rl.KeyUp):
			g.currentInput = key.command
		case int(rl.KeyDown):
			g.currentInput = key.command
		case int(rl.KeyLeft):
			g.currentInput = key.command
		case int(rl.KeyRight):
			g.currentInput = key.command
		case SpawnGopher:
			if num, err := strconv.Atoi(key.args[0]); err != nil {
				break
			} else {
				g.newGopher(num)
			}
		case KillGophs:
			g.destroyGophers()
		case Quack:
			if n, err := strconv.Atoi(key.args[0]); err != nil {
				break
			} else {
				g.quack(n)
			}
		case BigMouse:
			g.bigMouse = !g.bigMouse
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
				visuals.DisableMarquees()
				break
			}
			visuals.NewMarquee(key.args[0], float64(rand.Intn(250)+450), color.RGBA{0x00, 0xff, 0x00, 0xff}, false)
		case SingleMarqueeCmd:
			visuals.NewMarquee(key.args[0], float64(rand.Intn(250)+450), color.RGBA{0x00, 0xff, 0x00, 0xff}, true)
		case TTS:
			cache, _ := strconv.ParseBool(key.args[1])
			randomVoice, _ := strconv.ParseBool(key.args[2])
			go speech.Speak(key.args[0], cache, randomVoice)
		case BopCmd:
			g.bopometer.HandleMessage(key.args)
		case BingoCmd:
			g.bingoOverlay.HandleMessage(key.args)
		case MiracleCmd:
			g.showWhip = true
			sound.Play("indigo")
			go func() {
				time.Sleep(time.Second * 5)
				g.showWhip = false
			}()
		case MKCmd:
			g.showMK = true
			sound.Play("indigo")
			go func() {
				time.Sleep(time.Millisecond * 500)
				g.showMK = false
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
		case Quacksplosion:
			g.quacksplosion()
		case FollowAlert:
			if len(key.args) == 0 {
				break
			}
			visuals.ShowFollowAlert(key.args[0])
		case DedCmd:
			n, err := strconv.Atoi(key.args[0])
			if err != nil {
				break
			}
			dedCount = n
		case CubeCmd:
			go cube.HandleCommand(key.args)
		case TuxCmd:
			tuxpos.Z = -1000
			showtux = true
		case NowPlayingCmd:
			if key.args[0] == "off" {
				nowPlaying = ""
			} else {
				cps := getCodePointsFromString("Now Playing: " + key.args[0])
				fmt.Println(cps)
				ibmFont = rl.LoadFontEx("IBMPlexSansJP-Regular.otf", 48, cps)
				nowPlaying = key.args[0]
			}
		case NpTextCmd:
			if key.args[0] == "top" {
				npTextY = npTextTopY
				npBGY = 0
			} else if key.args[0] == "bottom" {
				npTextY = npTextBottomY
				npBGY = int32(npTextY - 10)
			}
		case MooCmd:
			sound.Play(moos[rand.Intn(len(moos))])
		case DropsCmd:
			fmt.Println("about to dro ps")
			visuals.ShowDrops(key.args[0])
		case DMCmd:
			g.showDM = !g.showDM
		case ToggleFSInfoCmd:
			g.showFSInfo = !g.showFSInfo
		case FSCmd:
			visuals.HandleFSCmd(key.args)
		case StreamCmd:
			if key.args[0] == "start" {
				startStream()
			} else if key.args[0] == "stop" {
				stopStream()
			} else if key.args[0] == "flip" {
				flipStreamCamera()
			}
		case MetricsCmd:
			if !visuals.MetricsEnabled() {
				visuals.EnableMetrics(true)
				speech.Speak("Metrics have arrived.", true, false)
			}
			lastMetricsUpdate = time.Now()
			visuals.HandleMetricsMessage(key.args)
		case RaidAlert:
			g.raidAlert()
		case SteamCmd:
			visuals.NewSteam().GetRandomGame()
		case GameCmd:
			games.HandleMessage(key.args)
		}
	default:
	}
	if g.gameRunning {
		g.snakeGame.Update(g.currentInput)
		g.currentInput = 0
	}
	games.Update(delta)
	if g.showStatic {
		g.staticLayer.Update()
	}
	g.bopometer.Update(delta)
	visuals.UpdateMarquees(delta)
	if g.showDM {
		visuals.UpdateDMarquee(delta)
	}
	if g.errorManager.Visible {
		g.errorManager.Update(delta)
	}

	for i := 0; i < g.sprites.num; i++ {
		if err := g.sprites.sprites[i].Update(delta); err != nil {
			return
		}
	}
	if visuals.MetricsEnabled() && time.Since(lastMetricsUpdate).Seconds() > 10 {
		go speech.Speak("Looks like I lost the metrics... Sorry about that.", true, false)
		visuals.EnableMetrics(false)
	}
	g.lastUpdate = time.Now()
}

func (g *Game) Draw() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.Color{R: 0x00, G: 0x00, B: 0x00, A: 0x00})
	rl.DrawPixel(0, 0, rl.Color{R: 0x00, G: 0x00, B: 0x00, A: 0x00})
	rl.BeginMode3D(camera)
	if showtux {
		rl.DrawBillboard(camera, sprites[2], tuxpos, 2.5, rl.White)
	}
	rl.EndMode3D()

	//	rl.DrawFPS(50, 50)

	g.errorManager.Draw()

	if dedCount > 0 {
		rl.DrawText(fmt.Sprintf("ded count: %d", dedCount), 25, 1340, 64, rl.Orange)
	}
	games.Draw()
	if g.showDM {
		visuals.DrawDMarquee()
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
	g.bopometer.Draw()
	cube.Draw()
	visuals.DrawDrops()
	visuals.DrawFollowAlert()
	if g.showFSInfo {
		visuals.DrawFSInfo()
	}
	visuals.DrawSteamOverlay()

	visuals.DrawMarquees()

	if g.showWhip {
		rl.DrawTextureEx(mwhipImg, rl.Vector2{X: 560, Y: 0}, 0, 0.6, rl.White)
	}

	if g.showMK {
		rl.DrawTextureEx(mkImg, rl.Vector2{X: 0, Y: 600}, 0, 1.0, rl.White)
	}

	g.bingoOverlay.Draw()

	visuals.DrawMetrics()

	if nowPlaying != "" {
		rl.DrawRectangle(0, npBGY, 2560, 75, rl.Color{R: 0, G: 0, B: 0, A: 192})
		rl.DrawTextEx(ibmFont, fmt.Sprintf("Now Playing: %s", nowPlaying), rl.Vector2{X: 25, Y: npTextY}, 48, 0, rl.SkyBlue)
	}

	rl.EndDrawing()
}

func main() {
	flag.Parse()

	http.HandleFunc("/go_pro_start", goProConnected)
	http.HandleFunc("/go_pro_stop", goProDisconnected)
	go http.ListenAndServe(":8083", nil)
	rl.SetConfigFlags(rl.FlagWindowMousePassthrough | rl.FlagWindowTopmost | rl.FlagWindowUndecorated | rl.FlagWindowTransparent)
	rl.InitWindow(screenWidth, screenHeight, "burtbot overlay")
	rl.SetTargetFPS(60)
	rl.InitAudioDevice()
	rl.SetMasterVolume(sound.MasterVolume)
	sound.LoadSounds()
	camera = rl.NewCamera3D(
		rl.Vector3{X: 0.0, Y: 0.0, Z: 10.0},
		rl.Vector3{X: 0.0, Y: 0.0, Z: 0.0},
		rl.Vector3{X: 0.0, Y: 1.0, Z: 0.0},
		45.0,
		rl.CameraPerspective,
	)
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	// if useANT {
	// 	usbCtx := gousb.NewContext()
	// 	defer usbCtx.Close()
	// 	startAntMonitor(usbCtx)
	// }

	mwhipImg = rl.LoadTexture("./images/mwhip.png")
	mkImg = rl.LoadTexture("./images/mk.png")
	carImg = rl.LoadTexture("./images/car.png")
	LoadSprites()
	shaders.LoadShaders()
	visuals.LoadFollowAlertAssets()
	visuals.LoadBopometerAssets()
	visuals.LoadDropsAssets()
	visuals.LoadFSAssets()
	visuals.InitMetrics()
	ga.commChannel = make(chan cmd)
	ga.connWriteChan = make(chan string)
	game := &ga
	game.bigMouseImg = sprites[2]
	visuals.LoadMarqueeFonts()
	goodFont = rl.LoadFontEx("caskaydia.TTF", 48, nil)
	ibmFont = rl.LoadFontEx("IBMPlexMono-Regular.ttf", 48, nil)
	npTextY = npTextBottomY
	npBGY = int32(npTextY - 10)
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
				//go speech.Speak("Intrusion Detected", true, false)
				conn.Close()
				continue
			}
			go handleConnection(conn, c, wc)
		}
	}(game.commChannel, ga.connWriteChan)
	games.Load(screenWidth, screenHeight, game.connWriteChan)
	defer games.Cleanup()
	game.snakeGame = newSnake()
	cube.LoadCubeAssets(game.connWriteChan)
	game.bopometer = visuals.NewBopometer(game.connWriteChan)
	game.bingoOverlay = visuals.NewBingoOverlay()
	game.errorManager = visuals.NewErrorManager()
	if err := visuals.PollFS(); err != nil {
		fmt.Println("Couldn't connect to sim")
	}

	goobsClient, err = goobs.New("localhost:4455", goobs.WithPassword(os.Getenv("OBSWS_PW")))
	if err != nil {
		log.Println("Couldn't connect to OBSWS - ", err.Error())
	}
	if goobsClient != nil {
		fmt.Println("Connected to OBSWS interface")
		defer goobsClient.Disconnect()
	}

	if showPlanes {
		fmt.Println("I've been asked to show planes")
		go func() {
			for {
				planes.CheckForPlanes()
				time.Sleep(time.Second * 5)
			}
		}()
	}

	for !rl.WindowShouldClose() {
		game.Update()
		game.Draw()
	}
	rl.CloseAudioDevice()
	rl.CloseWindow()
}

/*func startAntMonitor(ctx *gousb.Context) {
	usbDriver = ant.NewGarminStick3()
	scanner := ant.NewHeartRateScanner(usbDriver)
	scanner.ListenForData(func(s *ant.HeartRateScannerState) {
		if s.DeviceID == hrSensorID {
			currentHR = int(s.ComputedHeartRate)
		}
	})
	scanner.SetOnAttachCallback(func() {
		fmt.Println("Ant sensor attached")
	})
	usbDriver.OnStartup(func() {
		gettingHR = true
		scanner.Scan()
	})
	err := usbDriver.Open(ctx)
	if err != nil {
		log.Println("error opening usb driver: ", err.Error())
		usbDriver = nil
		return
	}
}*/

func handleConnection(conn net.Conn, c chan cmd, wc chan string) {
	defer conn.Close()
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		handleWrites(ctx, &conn, wc)
	}(ctx)
	defer cancel()
	fmt.Println("client connected")
	msg := connMessages[rand.Intn(len(connMessages))]
	go speech.Speak(msg, true, false)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Fields(txt)
		if len(fields) == 0 || fields[0] == "ping" {
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
		case "quack":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{Quack, fields[1:]}
		case "killgophs":
			c <- cmd{KillGophs, []string{}}
		case "bigmouse":
			c <- cmd{BigMouse, []string{}}
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
			if len(fields) < 4 {
				continue
			}
			c <- cmd{TTS, []string{strings.Join(fields[3:], " "), fields[1], fields[2]}}
		case "plinko":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{GameCmd, fields}
		case "tanks":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{GameCmd, fields}
		case "bop":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{BopCmd, fields[1:]}
		case "miracle":
			c <- cmd{MiracleCmd, []string{}}
		case "mk":
			c <- cmd{MKCmd, []string{}}
		case "lo":
			if len(fields) < 2 {
				continue
			}
			c <- cmd{GameCmd, append([]string{"lightsout"}, fields[1:]...)} //hacky
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
		case "quacksplosion":
			c <- cmd{Quacksplosion, []string{}}
		case "newfollow":
			c <- cmd{FollowAlert, fields[1:]}
		case "ded":
			c <- cmd{DedCmd, fields[1:]}
		case "cube":
			c <- cmd{CubeCmd, fields[1:]}
		case "tux":
			c <- cmd{TuxCmd, []string{}}
		case "nowplaying":
			c <- cmd{NowPlayingCmd, []string{strings.Join(fields[1:], " ")}}
		case "nptext":
			c <- cmd{NpTextCmd, fields[1:]}
		case "moo":
			c <- cmd{MooCmd, []string{}}
		case "itemdrops":
			c <- cmd{DropsCmd, []string{txt[10:]}}
		case "dmtoggle":
			c <- cmd{DMCmd, []string{}}
		case "fsToggle":
			c <- cmd{ToggleFSInfoCmd, []string{}}
		case "fs":
			c <- cmd{FSCmd, fields[1:]}
		case "stream":
			c <- cmd{StreamCmd, fields[1:]}
		case "hr":
			fallthrough
		case "cars":
			fallthrough
		case "speed":
			fallthrough
		case "distance":
			c <- cmd{MetricsCmd, fields}
		case "raidincoming":
			c <- cmd{RaidAlert, []string{}}
		case "steam":
			c <- cmd{SteamCmd, []string{}}
		case "slots":
			c <- cmd{GameCmd, fields}
		}

		if isVerbose {
			fmt.Println(fields)
		}
	}
}

func handleWrites(ctx context.Context, conn *net.Conn, wc chan string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Canceling tcp write loop")
			nowPlaying = ""
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
	if n <= 0 {
		return
	}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		if g.sprites.num == maxSprites {
			break
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

func (g *Game) quack(n int) {
	go func() {
		for i := 1; i <= n; i++ {
			sound.Play("quack")
			time.Sleep(time.Millisecond * 200)
		}
	}()
}

func (g *Game) quacksplosion() {
	go func() {
		var sleepTime = time.Duration(5000000000)
		for i := 1; i <= 30; i++ {
			time.Sleep(sleepTime)
			sound.Play("quack")
			fmt.Println(sleepTime)
			sleepTime = sleepTime / 2
			if sleepTime < 100000000 {
				sleepTime = 100000000
			}
		}
		sound.Play("explosion")
	}()
}

func (g *Game) raidAlert() {
	go func() {
		var sleepTime = time.Duration(500000000)
		for i := 1; i <= 5; i++ {
			sound.Play("voltage")
			time.Sleep(sleepTime)
		}
		speech.Speak(fmt.Sprintf("Sorry. This raid alert is broken. Please try again another time and apologies for the inconvenience. Error Number %d", time.Now().UnixNano()), true, false)
	}()
}

// start the stream
func startStream() bool {
	if obsCmd != nil {
		log.Println("obs is already running")
		return false
	}
	fmt.Println("attempting stream start")
	cmd := exec.Command("obs", "--scene", "outdoors", "--startstreaming")
	if err := cmd.Start(); err != nil {
		log.Println(err)
		return false
	}
	obsCmd = cmd
	log.Println(fmt.Sprintf("Started obs with PID %d", cmd.Process.Pid))
	return true
}

func stopStream() {
	if obsCmd == nil {
		return
	}
	obsCmd.Process.Kill()
	obsCmd = nil
}

// TODO - Need to account for the scenario where the overlay closes
// mid stream. As of now, the only way to pick back up the stream
// health checks would be for the go pro stream to restart to
// trigger this event.  Maybe on overlay start, do one health check
// and if we find a publisher, call this function manually??
func goProConnected(w http.ResponseWriter, r *http.Request) {
	speech.Speak("Go pro connected", true, false)
	if goobsClient != nil {
		params := scenes.NewSetCurrentProgramSceneParams().
			WithSceneName("outdoors")
		goobsClient.Scenes.SetCurrentProgramScene(params)
	}
	if streamHealthCancelFunc != nil {
		// if we already have a health check running and
		// it didn't close properly, don't start another
		fmt.Fprintf(w, "got it\n")
		return
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	streamHealthCancelFunc = cancelFunc
	go func(ctx context.Context) {
		fmt.Println("starting health check llooolp")
		ticker := time.NewTicker(time.Second * 3)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fmt.Println("tick")
				checkGoProStreamHealth()
			}
		}
	}(ctx)
	fmt.Fprintf(w, "got it\n")
}

func goProDisconnected(w http.ResponseWriter, r *http.Request) {
	speech.Speak("Go pro disconnected", true, false)
	if streamHealthCancelFunc != nil {
		// Clean up any stream health checks
		params := scenes.NewSetCurrentProgramSceneParams().
			WithSceneName("no_signal")
		goobsClient.Scenes.SetCurrentProgramScene(params)
		streamHealthCancelFunc()
	}
	fmt.Fprintf(w, "go it\n")
}

func checkGoProStreamHealth() {
	respStruct := struct {
		Apps []RTMPApplication `xml:"server>application"`
	}{}

	req, err := http.NewRequest("GET", "http://192.168.0.29:8080/stat", nil)
	if err != nil {
		log.Println("Error getting stream health update", err.Error())
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error getting stream health", err.Error())
		return
	}

	err = xml.NewDecoder(resp.Body).Decode(&respStruct)
	if err != nil {
		fmt.Println("Err parsing resp body xml health", err.Error())
		return
	}

	fmt.Println("steam heltj chekc sone")
	// Check to see if we have any active streams
	// If not, change scene to our lost signal stream
	// we can check for application outdoor having a client with state: publishing
	for _, app := range respStruct.Apps {
		if app.Name == "outdoor" {
			if len(app.Publishing) == 0 {
				// No publishing - lost signal
				// should be covered by go pro disc event?
				fmt.Println("No stream publisher found - ??")
				continue
			}
			if app.BitRate < 1800000 {
				//Low bitrate - maybe change scenes
				// or show low bitrate warning, see what we can
				// do over websocket
				fmt.Println("---- Stream below bitrate threshold -----", app.BitRate)
			}
			fmt.Printf("Outdoor bitrate: %d\n", app.BitRate)
		}
	}
}

func flipStreamCamera() {

	/*prams := sceneitems.NewGetSceneItemListParams().
	      WithSceneName("outdoors")
	  r, err := goobsClient.SceneItems.GetSceneItemList(prams)
	  if err != nil {
	      fmt.Println("error getting scene item list", err.Error())
	      return
	  }
	  for _, v := range r.SceneItems {
	      fmt.Printf("Scene ID: %d - Source Name: %s\n", v.SceneItemID, v.SourceName)
	  }*/
	gsitParams := sceneitems.NewGetSceneItemTransformParams().
		WithSceneName("outdoors").
		WithSceneItemId(1)
	gsitResp, err := goobsClient.SceneItems.GetSceneItemTransform(gsitParams)
	if err != nil {
		fmt.Println("Error getting transform ", err.Error())
		return
	}
	newTransform := gsitResp.SceneItemTransform
	newTransform.Rotation = float64(int(newTransform.Rotation+180) % 360)
	params := sceneitems.NewSetSceneItemTransformParams().
		WithSceneName("outdoors").
		WithSceneItemId(1).
		WithSceneItemTransform(newTransform)
	_, err = goobsClient.SceneItems.SetSceneItemTransform(params)
	if err != nil {
		fmt.Println("Error setting transform ", err.Error())
	}
}

func getCodePointsFromString(s string) []rune {
	codePoints := []rune{}
	for _, r := range s {
		codePoints = append(codePoints, r)
	}
	return codePoints
}

// perform any necessary cleanup here. should be called on
// a interrupt signal and any other form of exit
func cleanUp() {
	fmt.Println("cleaning up after forcful exit")
	// if usbDriver != nil {
	// 	usbDriver.Close()
	// }
	// rl.CloseAudioDevice()
	// rl.CloseWindow()
	if streamHealthCancelFunc != nil {
		streamHealthCancelFunc()
	}
	fmt.Println("save the cube!")
	cube.SaveCube()
	os.Exit(0)
}
