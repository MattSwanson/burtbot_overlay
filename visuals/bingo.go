package visuals

import (
	"fmt"
	"time"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

const (
	numberMemory = 5
)

type BingoOverlay struct {
	currentNumber   string
	previousNumbers []string
	winner          string
	prize           string
	gameOver        bool
	display         bool
}

func NewBingoOverlay() *BingoOverlay {
	return &BingoOverlay{
		previousNumbers: make([]string, numberMemory),
	}
}

func (b *BingoOverlay) AddNumber(num string) {
	b.currentNumber, b.previousNumbers = num, append(b.previousNumbers[1:], b.currentNumber)
}

func (b *BingoOverlay) Draw() {
	// draw the last drawn number
	const rightEdge int32 = 2460
	cnts := rl.MeasureText(b.currentNumber, 96)
	tws := make([]int32, 5)
	for i, str := range b.previousNumbers {
		tws[i] = rl.MeasureText(str, 40+int32(i)*8)
	}
	rl.DrawText(b.currentNumber, rightEdge-cnts, 1244, 96, rl.SkyBlue)
	rl.DrawText(b.previousNumbers[4], rightEdge-tws[4], 1169, 72, rl.LightGray)
	rl.DrawText(b.previousNumbers[3], rightEdge-tws[3], 1102, 64, rl.LightGray)
	rl.DrawText(b.previousNumbers[2], rightEdge-tws[2], 1045, 56, rl.LightGray)
	rl.DrawText(b.previousNumbers[1], rightEdge-tws[1], 998, 48, rl.LightGray)
	rl.DrawText(b.previousNumbers[0], rightEdge-tws[0], 961, 40, rl.LightGray)
	// draw the previous 3 numbers to the

	// if game ended draw winner name and profile image
	if b.gameOver {
		s := fmt.Sprintf("%s has BINGO! They won %s tokens!", b.winner, b.prize)
		rl.DrawText(s, 100, 1175, 96, rl.Lime)
	}
}

func (b *BingoOverlay) Reset() {
	b.currentNumber = ""
	b.previousNumbers = make([]string, 5)
}

func (b *BingoOverlay) End(username, prize string) {
	b.winner, b.prize = username, prize
	b.gameOver = true
	go func() {
		time.Sleep(time.Second * 5)
		b.gameOver = false
	}()
}
