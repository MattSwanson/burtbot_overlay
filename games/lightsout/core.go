package lightsout

import (
	"fmt"
	"strconv"
	"time"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

const (
	gameHeight = 1000
	gameWidth  = 1000
)

type Core struct {
	gameBoard      []*light
	numColumns     int
	numRows        int
	currentPuzzle  int
	puzzleComplete bool
	running        bool
}

var puzzles = [][]int{
	{
		1, 1, 1, 1, 1,
		1, 1, 0, 1, 1,
		1, 0, 0, 0, 1,
		1, 1, 0, 1, 1,
		1, 1, 1, 1, 1,
	},
	{
		0, 0, 1, 0, 0,
		0, 1, 1, 1, 0,
		1, 1, 1, 1, 1,
		0, 1, 1, 1, 0,
		0, 0, 1, 0, 0,
	},
	{
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
	},
}

func NewGame(w, h int) *Core {
	gameBoard := []*light{}
	leftEdge := (rl.GetScreenWidth() - gameWidth) / 2
	topEdge := (rl.GetScreenHeight() - gameHeight) / 2
	lightWidth := gameWidth / w
	lightHeight := gameHeight / h
	for i := 0; i < w*h; i++ {
		x := leftEdge + lightWidth*(i%w)
		y := topEdge + lightHeight*(i/h)
		l := NewLight(int32(x), int32(y), int32(lightWidth), int32(lightHeight))
		gameBoard = append(gameBoard, l)
	}
	return &Core{
		gameBoard:  gameBoard,
		numColumns: w,
		numRows:    h,
	}
}

func (c *Core) HandleMessage(args []string) {
	if args[0] == "start" && !c.running {
		c.LoadPuzzle(0)
		c.running = true
		return
	}
	if !c.running {
		return
	}
	if args[0] == "reset" {
		c.Reset()
		return
	}
	if args[0] == "stop" {
		c.running = false
		return
	}
	n, err := strconv.Atoi(args[0])
	if err != nil {
		return
	}
	c.Press(n)
}

func (c *Core) Reset() {
	c.puzzleComplete = false
	for _, l := range c.gameBoard {
		l.on = false
	}
}

func (c *Core) LoadPuzzle(i int) {
	c.Reset()
	for k, n := range puzzles[i] {
		if n == 1 {
			c.gameBoard[k].on = true
		}
	}
}

func (c *Core) Press(pos int) {
	if pos >= c.numColumns*c.numRows || pos < 0 {
		return
	}
	c.gameBoard[pos].Toggle()
	// Then toggle adjacent lights
	// up
	if pos/c.numRows != 0 {
		c.gameBoard[pos-c.numColumns].Toggle()
	}
	// left
	if pos%c.numColumns != 0 {
		c.gameBoard[pos-1].Toggle()
	}
	// down
	if pos/c.numRows < c.numRows-1 {
		c.gameBoard[pos+c.numColumns].Toggle()
	}
	// right
	if pos%c.numColumns != c.numColumns-1 {
		c.gameBoard[pos+1].Toggle()
	}
	if CheckForWin(c.gameBoard) {
		// wind screen
		c.puzzleComplete = true
		go func() {
			time.Sleep(time.Second * 10)
			c.currentPuzzle++
			if c.currentPuzzle >= len(puzzles) {
				// Gug?
				fmt.Println("add more puzzles scrub. yeah you.")
			} else {
				c.LoadPuzzle(c.currentPuzzle)
			}
		}()
	}
}

func CheckForWin(board []*light) bool {
	for _, l := range board {
		if !l.on {
			return false
		}
	}
	return true
}

func (c *Core) Draw() {
	if !c.running {
		return
	}
	for k, l := range c.gameBoard {
		l.Draw()
		// x, y := k%c.numColumns, k/c.numRows
		rl.DrawText(fmt.Sprint(k), l.x, l.y, 18, rl.Green)
	}
	if c.puzzleComplete {
		rl.DrawText("winrar", 400, 500, 96, rl.Color{R: 0x55, G: 0xBA, B: 0x2C, A: 0xFF})
	}
}

type light struct {
	x  int32
	y  int32
	w  int32
	h  int32
	on bool
}

func NewLight(x, y, w, h int32) *light {
	return &light{x, y, w, h, false}
}

func (l *light) Toggle() {
	l.on = !l.on
}

func (l *light) Draw() {
	if l.on {
		rl.DrawRectangle(l.x, l.y, l.w, l.h, rl.Red)
	}
}
