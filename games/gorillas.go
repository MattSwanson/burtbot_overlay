package games

import (
	"github.com/MattSwanson/ebiten/v2"
)

// @velusip's idea
type Gorillas struct {
	currentPlayer  int
	scoreLimit     int
	playerOne      string
	playerOneScore int
	playerTwo      string
	playerTwoScore int
	windVelocity   float64 // - wind push right + push left
	blastRadius    float64
	banana         banana
}

type banana struct {
	inFlight bool
	x        float64
	y        float64
	vx       float64
	vy       float64
}

type gorilla struct {
	name string
	x    float64
	y    float64
}

func init() {
	// load image for banana
	// load image for gorillas
}

func (g *Gorillas) ResetGame() {
	g.currentPlayer = 1
	g.playerOneScore = 0
	g.playerTwoScore = 0
	g.windVelocity = 0
}

func (g *Gorillas) Update() error {
	if g.banana.inFlight {
		g.banana.x += g.banana.vx
		g.banana.y += g.banana.vy
	}
	return nil
}

func (g *Gorillas) Draw(screen *ebiten.Image) {

}

func (g *Gorillas) CreateCity() {

}
