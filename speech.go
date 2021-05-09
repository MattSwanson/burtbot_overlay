package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"strings"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	rl "github.com/MattSwanson/raylib-go/raylib"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

const (
	ttsSampleRate = 44100
)

var cache []string // a slice of strings representing the sha256 of cached tts, loaded at init based on files present

func init() {
	cache = []string{}
	files, err := os.ReadDir("tts_cache")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fmt.Println(file.Name())
		cache = append(cache, file.Name())
	}
}

func speak(txt string, shouldCache bool) error {
	var hash string
	var cached bool
	if shouldCache {
		txt = strings.ToLower(txt)
		h := sha256.New()
		h.Write([]byte(txt))
		hash = fmt.Sprintf("%x", h.Sum(nil))
		fmt.Println(hash)
		for _, s := range cache {
			if s == hash+".wav" {
				fmt.Println("we cached")
				cached = true
			}
		}
	}
	var audioBytes []byte
	var err error
	var sound rl.Sound
	if !cached {
		audioBytes, err = getTTS(txt)
		if err != nil {
			log.Println("Couldn't get TTS: ", err.Error())
			return err
		}
		if shouldCache {
			err = os.WriteFile(fmt.Sprintf("tts_cache/%s.wav", hash), audioBytes, 0666)
			if err != nil {
				log.Println("unable to write wave to file", err.Error())
			}
		}
		wave := rl.NewWave(uint32(len(audioBytes)/2), ttsSampleRate, 16, 1, audioBytes[44:])
		sound = rl.LoadSoundFromWave(wave)
	} else {
		sound = rl.LoadSound(fmt.Sprintf("tts_cache/%s.wav", hash))
	}
	rl.PlaySoundMulti(sound)
	return nil
}

func getTTS(txt string) ([]byte, error) {
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Printf("Unable to start tts client: %s\n", err)
		return nil, err
	}

	req := texttospeechpb.SynthesizeSpeechRequest{
		// set the text input to be synthesized
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: txt},
		},
		// Build the voice request, select the language code ("en-us") and the SSML
		// voice gender
		// en-US-Wavenet-E or en-US-Wavenet-J are top picks
		Voice: &texttospeechpb.VoiceSelectionParams{
			Name:         "en-US-Wavenet-J",
			LanguageCode: "en-US",
			//SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL,
		},
		// select the type of audio you want returned
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding:   texttospeechpb.AudioEncoding_LINEAR16,
			SampleRateHertz: 44100,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		log.Printf("Couldn't synthesize speech: %s\n", err)
		return nil, err
	}

	return resp.AudioContent, nil
}

func getAvailableVoices() (string, error) {
	ctx := context.Background()
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Printf("Unable to start tts client: %s\n", err)
		return "", err
	}
	lvRequest := texttospeechpb.ListVoicesRequest{LanguageCode: "en-US"}
	resp, err := client.ListVoices(ctx, &lvRequest)
	if err != nil {
		log.Printf("Unable to get the list of voices from the API: %s", err.Error())
		return "", err
	}
	for _, v := range resp.Voices {
		fmt.Println(v.String())
	}
	return resp.String(), nil
}
