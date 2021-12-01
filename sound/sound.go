package sound

import (
	"errors"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

const MasterVolume float32 = 0.75

var sounds map[string]rl.Sound = map[string]rl.Sound{}

func LoadSounds() {
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
	sounds["explosion"] = rl.LoadSound("sounds/explosion-02.wav")
	sounds["moo_a1"] = rl.LoadSound("sounds/moo/attack1.wav")
	sounds["moo_a2"] = rl.LoadSound("sounds/moo/attack2.wav")
	sounds["moo_a3"] = rl.LoadSound("sounds/moo/attack3.wav")
	sounds["moo_a4"] = rl.LoadSound("sounds/moo/attack4.wav")
	sounds["moo_a5"] = rl.LoadSound("sounds/moo/attack5.wav")
	sounds["moo_a6"] = rl.LoadSound("sounds/moo/attack6.wav")
	sounds["moo_d1"] = rl.LoadSound("sounds/moo/death1.wav")
	sounds["moo_d2"] = rl.LoadSound("sounds/moo/death2.wav")
	sounds["moo_d3"] = rl.LoadSound("sounds/moo/death3.wav")
	sounds["moo_d4"] = rl.LoadSound("sounds/moo/death4.wav")
	sounds["moo_d5"] = rl.LoadSound("sounds/moo/death5.wav")
	sounds["moo_h1"] = rl.LoadSound("sounds/moo/gethit1.wav")
	sounds["moo_h2"] = rl.LoadSound("sounds/moo/gethit2.wav")
	sounds["moo_h3"] = rl.LoadSound("sounds/moo/gethit3.wav")
	sounds["moo_h4"] = rl.LoadSound("sounds/moo/gethit4.wav")
	sounds["moo_n1"] = rl.LoadSound("sounds/moo/neutral1.wav")
	sounds["moo_n2"] = rl.LoadSound("sounds/moo/neutral2.wav")
	sounds["moo_n3"] = rl.LoadSound("sounds/moo/neutral3.wav")
	sounds["moo_n4"] = rl.LoadSound("sounds/moo/neutral4.wav")
	sounds["moo_n5"] = rl.LoadSound("sounds/moo/neutral5.wav")

}

func Play(name string) error {
	if _, ok := sounds[name]; !ok {
		return errors.New("sound is not loaded")
	}
	rl.PlaySoundMulti(sounds[name])
	return nil
}
