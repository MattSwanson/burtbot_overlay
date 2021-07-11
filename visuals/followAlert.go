package visuals

import (
	"fmt"
	"time"

	"github.com/MattSwanson/burtbot_overlay/sound"
	rl "github.com/MattSwanson/raylib-go/raylib"
)

const (
	userNameTextXCenter = 1675
	alertLength         = 6 // seconds
)

var (
	alertVisible   bool
	largeGopher    rl.Texture2D
	speechBubble   rl.Texture2D
	userNamePosX   int32 = 900
	userNameString string
)

func LoadFollowAlertAssets() {
	largeGopher = rl.LoadTexture("./images/tux_goph.png")
	speechBubble = rl.LoadTexture("./images/speech_bubble.png")
}

func ShowFollowAlert(username string) {
	fmt.Println("new follower: ", username)
	sound.Play("eep")
	userNameString = fmt.Sprintf("%s!", username)
	textWidth := rl.MeasureText(userNameString, 96)
	fmt.Println(textWidth)
	userNamePosX = userNameTextXCenter - textWidth/2
	fmt.Println(userNamePosX)
	alertVisible = true
	go func() {
		time.Sleep(time.Second * 10)
		alertVisible = false
	}()
}

func DrawFollowAlert() {
	if !alertVisible {
		return
	}

	// draw large gopher
	rl.DrawTexture(largeGopher, -200, 0, rl.White)
	rl.DrawTexture(speechBubble, 675, 200, rl.White)
	// draw text with message and user name
	rl.DrawText("Thanks for following,", 1000, 450, 96, rl.DarkBlue)
	rl.DrawText(userNameString, userNamePosX, 550, 96, rl.Orange)
}
