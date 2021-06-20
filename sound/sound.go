package sound

import (
	"errors"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

var sounds map[string]rl.Sound

const MasterVolume float32 = 0.75

func init() {
	sounds = map[string]rl.Sound{}
	sounds["eep"] = rl.LoadSound("sounds/wildeep.wav")
	sounds["whit"] = rl.LoadSound("sounds/Whit.wav")
	sounds["boing"] = rl.LoadSound("sounds/Boing.wav")
	sounds["quack"] = rl.LoadSound("sounds/Quack.wav")
	sounds["zap"] = rl.LoadSound("sounds/Voltage.wav")
	sounds["logjam"] = rl.LoadSound("sounds/Logjam.wav")
	sounds["bip"] = rl.LoadSound("sounds/Bip.wav")
	sounds["squeek"] = rl.LoadSound("sounds/ChuToy.wav")
	sounds["indigo"] = rl.LoadSound("sounds/Indigo.wav")
	sounds["sosumi"] = rl.LoadSound("sounds/Sosumi.wav")
	sounds["kerplunk"] = rl.LoadSound("sounds/kerplunk.wav")
}

func Play(name string) error {
	if _, ok := sounds[name]; !ok {
		return errors.New("sound is not loaded")
	}
	rl.PlaySoundMulti(sounds[name])
	return nil
}
