package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"github.com/MattSwanson/ebiten/v2/audio"
	"github.com/MattSwanson/ebiten/v2/audio/mp3"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

const (
	ttsSampleRate = 44100
)

func speak(audioContext *audio.Context, txt string) error {
	audioBytes, err := getTTS(txt)
	if err != nil {
		log.Println("Couldn't get TTS: ", err.Error())
		return err
	}
	ws, err := mp3.DecodeWithSampleRate(ttsSampleRate, bytes.NewReader(audioBytes))
	if err != nil {
		return err
	}
	player, err := audio.NewPlayer(audioContext, ws)
	if err != nil {
		return err
	}
	player.Play()
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
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
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
