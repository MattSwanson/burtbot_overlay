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
	followTextSize      = 96
)

var (
	alertVisible   bool
	largeGopher    rl.Texture2D
	speechBubble   rl.Texture2D
	userNamePosX   int32 = 900
	userNameString string
	followFont     rl.Font
)

func LoadFollowAlertAssets() {
	largeGopher = rl.LoadTexture("./images/tux_goph.png")
	speechBubble = rl.LoadTexture("./images/speech_bubble.png")
	followFont = rl.LoadFontEx("./caskaydia.TTF", followTextSize, nil, 0)
}

func ShowFollowAlert(username string) {
	fmt.Println("new follower: ", username)
	sound.Play("eep")
	userNameString = fmt.Sprintf("%s!", username)
	textWidth := rl.MeasureTextEx(followFont, userNameString, followTextSize, 0).X
	fmt.Println(textWidth)
	userNamePosX = userNameTextXCenter - int32(textWidth/float32(2))
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
	rl.DrawTextEx(followFont, "Thanks for following,", rl.Vector2{X: 1000, Y: 450}, followTextSize, 0, rl.DarkBlue)
	rl.DrawTextEx(followFont, userNameString, rl.Vector2{X: float32(userNamePosX), Y: 550.0}, followTextSize, 0, rl.Orange)
}
