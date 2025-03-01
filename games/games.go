package games

import (
	"fmt"

	"github.com/MattSwanson/burtbot_overlay/games/lightsout"
	"github.com/MattSwanson/burtbot_overlay/games/plinko"
	"github.com/MattSwanson/burtbot_overlay/games/slots"
	"github.com/MattSwanson/burtbot_overlay/games/tanks"
)

type Game interface {
	Cleanup()
	Draw()
	HandleMessage([]string)
	Update(float64)
}

var games map[string]Game = map[string]Game{}
var drawOrder []string = []string{
	"slots",
	"tanks",
	"lightsout",
	"plinko",
}

func Load(screenWidth, screenHeight float64, writeChannel chan string) {
	games["plinko"] = plinko.Load(screenWidth, screenHeight, writeChannel)
	games["tanks"] = tanks.Load(screenWidth, screenHeight)
	games["lightsout"] = lightsout.NewGame(5, 5)
	games["slots"] = slots.LoadSlots(writeChannel)
}

func Draw() {
	for _, key := range drawOrder {
		games[key].Draw()
	}
}

// Update the games state
func Update(delta float64) {
	for _, game := range games {
		game.Update(delta)
	}
}

// First element in the slice should be the name of the
// game we want to send a message to. If not in the map
// then ignored
func HandleMessage(message []string) {
	fmt.Println("handlign thr mnsase", message[0])
	if game, ok := games[message[0]]; ok {
		game.HandleMessage(message[1:])
	}
}

func Cleanup() {
	for _, game := range games {
		game.Cleanup()
	}
}
