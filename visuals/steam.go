package visuals

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"math/rand"
	"net/http"
	"os"
	"time"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

const (
	timePlayedThreshold = 10
	timerStart          = 60 * 60
)

var userID string = "76561197968481769"
var draw bool
var drawTimer bool
var appImg rl.Texture2D
var gameName string
var bgWidth int
var timeRemaining int = timerStart

type Steam struct {
}

type appEntry struct {
	AppID           int    `json:"appid"`
	PlaytimeForever int    `json:"playtime_forever"`
	TimeLastPlayed  int    `json:"rtime_last_played"`
	Name            string `json:"name"`
	ImgIconURL      string `json:"img_icon_url"`
	IconTex         rl.Texture2D
}

type steamAPIResponse struct {
	Response struct {
		Games []appEntry `json:"games"`
	} `json:"response"`
}

func NewSteam() *Steam {
	return &Steam{}
}

func DrawSteamOverlay() {
	if !draw {
		return
	}
	rl.DrawRectangleV(rl.Vector2{float32(screenWidth)/3 - 20, float32(screenHeight)/3 - 20}, rl.Vector2{float32(bgWidth), 272}, rl.Color{0, 0, 0, 192})
	rl.DrawTextureEx(appImg, rl.Vector2{float32(screenWidth) / 3, float32(screenHeight) / 3}, 0, 5.0, rl.White)
	rl.DrawTextEx(bopFont, gameName, rl.Vector2{float32(screenWidth) / 3, float32(screenHeight)/3 + 170}, 72.0, 0, rl.Blue)

}

func (s *Steam) GetRandomGame() {
	apiKey := os.Getenv("STEAM_API_KEY")
	url := fmt.Sprintf("http://api.steampowered.com/IPlayerService/GetOwnedGames/v0001/?key=%s&steamid=%s&format=json&include_appinfo=1", apiKey, userID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Steam api err: ", err.Error())
		return
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error accessing Steam API: ", err.Error())
		return
	}
	r := steamAPIResponse{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	filtered := []appEntry{}
	for _, game := range r.Response.Games {
		if game.PlaytimeForever < timePlayedThreshold {
			filtered = append(filtered, game)
		}
	}

	// shuffle the games
	rand.Shuffle(len(filtered), func(i, j int) {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	})
	fmt.Println("set is ", len(filtered))
	filtered = filtered[:20]

	// Get the first 20? icons
	var img []image.Image = make([]image.Image, 20)
	var rimg []*rl.Image = make([]*rl.Image, 20)
	for k, app := range filtered {
		url = fmt.Sprintf("https://media.steampowered.com/steamcommunity/public/images/apps/%d/%s.jpg", app.AppID, app.ImgIconURL)
		resp, err = http.Get(url)
		if err != nil {
			fmt.Println("Couldn't get img for steam api ", err.Error())
			return
		}
		img[k], _, err = image.Decode(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
		}
		recolor := image.NewRGBA(image.Rect(0, 0, img[k].Bounds().Dx(), img[k].Bounds().Dy()))
		for x := 0; x < img[k].Bounds().Dx(); x++ {
			for y := 0; y < img[k].Bounds().Dy(); y++ {
				recolor.Set(x, y, img[k].At(x, y))
			}
		}
		rimg[k] = rl.NewImageFromImage(recolor)
		filtered[k].IconTex = rl.LoadTextureFromImage(rimg[k])
		resp.Body.Close()
	}

	appImg = filtered[0].IconTex
	gameName = ""
	bgWidth = 0
	draw = true
	go func() {
		randOff := rand.Intn(40) - 20
		winner := 0
		for i := 0; i < 100-randOff; i++ {
			idx := i % len(filtered)
			appImg = filtered[idx].IconTex
			winner = idx
			time.Sleep(time.Millisecond * 70)
		}
		gameName = filtered[winner].Name
		bgWidth = int(rl.MeasureTextEx(bopFont, gameName, 72, 0).X) + 40
		time.Sleep(time.Second * 30)
		draw = false
	}()
}
