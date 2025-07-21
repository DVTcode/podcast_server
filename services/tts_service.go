package services

import (
	"context"
	"errors"
	"os"
	
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

// SynthesizeText chuyển text thành giọng nói
func SynthesizeText(text string, voice string, rate float64) ([]byte, error) {
	if len(text) == 0 {
		return nil, errors.New("text is empty")
	}
	if voice == "" {
		voice = "vi-VN-Chirp3-HD-Puck"
	}
	if rate <= 0 {
		rate = 1.0
	}

	ctx := context.Background()
	
	jsonCreds := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if jsonCreds == "" {
		return nil, errors.New("GOOGLE_CREDENTIALS_JSON environment variable is not set")
	}

	client, err := texttospeech.NewClient(ctx, option.WithCredentialsJSON([]byte(jsonCreds)))
	
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var allAudio []byte
	chunks := splitTextToChunks(text, 4500) // nhỏ hơn 5000 bytes để an toàn

	for _, chunk := range chunks {
		req := &texttospeechpb.SynthesizeSpeechRequest{
			Input: &texttospeechpb.SynthesisInput{
				InputSource: &texttospeechpb.SynthesisInput_Text{
					Text: chunk,
				},
			},
			Voice: &texttospeechpb.VoiceSelectionParams{
				LanguageCode: "vi-VN",
				Name:         voice,
			},
			AudioConfig: &texttospeechpb.AudioConfig{
				AudioEncoding: texttospeechpb.AudioEncoding_MP3,
				SpeakingRate:  rate,
			},
		}

		resp, err := client.SynthesizeSpeech(ctx, req)
		if err != nil {
			return nil, err
		}
		allAudio = append(allAudio, resp.AudioContent...)
	}

	return allAudio, nil
}

func splitTextToChunks(text string, maxLen int) []string {
	var chunks []string
	runes := []rune(text)
	for len(runes) > 0 {
		if len(runes) > maxLen {
			chunks = append(chunks, string(runes[:maxLen]))
			runes = runes[maxLen:]
		} else {
			chunks = append(chunks, string(runes))
			break
		}
	}
	return chunks
}
