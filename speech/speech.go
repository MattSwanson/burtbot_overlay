package speech

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	rl "github.com/MattSwanson/raylib-go/raylib"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

const (
	ttsSampleRate            = 44100
	defaultVoiceName         = "en-US-Wavenet-J"
	defaultVoiceLanguageCode = "en-US"
	defaultVoiceSampleRate   = 44100
)

var cache []string // a slice of strings representing the sha256 of cached tts, loaded at init based on files present
var voices []*texttospeechpb.Voice
var currentSampleRate int32

func init() {
	cache = []string{}
	files, err := os.ReadDir("tts_cache")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		cache = append(cache, file.Name())
	}
}

func Speak(txt string, shouldCache, useRandomVoice bool) error {

	var hash string
	var cached bool
	if shouldCache {
		txt = strings.ToLower(txt)
		h := sha256.New()
		h.Write([]byte(txt))
		hash = fmt.Sprintf("%x", h.Sum(nil))
		for _, s := range cache {
			if s == hash+".wav" {
				cached = true
			}
		}
	}
	var audioBytes []byte
	var err error
	var sound rl.Sound
	if !cached {
		audioBytes, err = getTTS(txt, useRandomVoice)
		if err != nil {
			log.Println("Couldn't get TTS: ", err.Error())
			return err
		}
		if shouldCache {
			filename := fmt.Sprintf("tts_cache/%s.wav", hash)
			err = os.WriteFile(filename, audioBytes, 0666)
			if err != nil {
				log.Println("unable to write wave to file", err.Error())
			} else {
				cache = append(cache, filename)
			}
		}
		wave := rl.NewWave(uint32(len(audioBytes)/2), uint32(currentSampleRate), 16, 1, audioBytes[44:])
		// garbage := []byte{}
		// for i := 0; i < 100000; i++ {
		// 	garbage = append(garbage, byte(rand.Intn(256)))
		// }
		//garbageWave := rl.NewWave(uint32(len(garbage)/2), 11000, 16, 1, garbage)
		//garbageSound := rl.LoadSoundFromWave(garbageWave)
		sound = rl.LoadSoundFromWave(wave)
		//rl.PlaySoundMulti(garbageSound)
	} else {
		sound = rl.LoadSound(fmt.Sprintf("tts_cache/%s.wav", hash))
	}
	if rl.GetSoundsPlaying() >= 16 {
		fmt.Println("TOO MUCH SOUNDS PLZ SOPT")
		// queue this sound up to be played when able?
		go func(sound rl.Sound) {
			for {
				time.Sleep(time.Millisecond * 50)
				if rl.GetSoundsPlaying() < 16 {
					rl.PlaySoundMulti(sound)
					break
				}
			}
		}(sound)
	}
	rl.PlaySoundMulti(sound)
	return nil
}

func getTTS(txt string, useRandomVoice bool) ([]byte, error) {
	if len(voices) < 1 {
		var err error
		voices, err = getAvailableVoices()
		if err != nil {
			return []byte{}, err
		}
	}
	ctx, canc := context.WithTimeout(context.Background(), time.Second*10)
	defer canc()
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Printf("Unable to start tts client: %s\n", err)
		return nil, err
	}
	defer client.Close()

	//speed := rand.Float64() + 0.5

	req := texttospeechpb.SynthesizeSpeechRequest{
		// set the text input to be synthesized
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: txt},
		},
		// Build the voice request, select the language code ("en-us") and the SSML
		// voice gender
		// en-US-Wavenet-E or en-US-Wavenet-J are top picks
		Voice: &texttospeechpb.VoiceSelectionParams{
			Name:         defaultVoiceName,
			LanguageCode: defaultVoiceLanguageCode,
		},
		// select the type of audio you want returned
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding:   texttospeechpb.AudioEncoding_LINEAR16,
			SampleRateHertz: defaultVoiceSampleRate,
			SpeakingRate:    1.0,
		},
	}
	currentSampleRate = defaultVoiceSampleRate

	if useRandomVoice {
		n := rand.Intn(len(voices))
		req.Voice.Name = voices[n].Name
		req.Voice.LanguageCode = voices[n].LanguageCodes[0]
		req.Voice.SsmlGender = voices[n].SsmlGender
		req.AudioConfig.SampleRateHertz = voices[n].NaturalSampleRateHertz
		currentSampleRate = voices[n].NaturalSampleRateHertz
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		log.Printf("Couldn't synthesize speech: %s\n", err)
		return nil, err
	}

	return resp.AudioContent, nil
}

func getAvailableVoices() ([]*texttospeechpb.Voice, error) {
	ctx := context.Background()
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Printf("Unable to start tts client: %s\n", err)
		return []*texttospeechpb.Voice{}, err
	}
	lvRequest := texttospeechpb.ListVoicesRequest{}
	resp, err := client.ListVoices(ctx, &lvRequest)
	if err != nil {
		log.Printf("Unable to get the list of voices from the API: %s", err.Error())
		return []*texttospeechpb.Voice{}, err
	}
	return resp.Voices, nil
}
