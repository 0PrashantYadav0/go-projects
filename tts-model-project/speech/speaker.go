package speech

import (
	"context"
	"fmt"
	"io"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

var speaker *Speaker

const (
	//ref http://cloud.google.com/text-to-speech/docs/voices
	VoiceStandardA         = "ja-JP-Standard-A"
	VoiceStandardB         = "ja-JP-Standard-B"
	VoiceStandardC         = "ja-JP-Standard-C"
	VoiceStandardD         = "ja-JP-Standard-D"
	VoiceWavenetA          = "ja-JP-Wavenet-A"
	VoiceWavenetB          = "ja-JP-Wavenet-B"
	VoiceWavenetC          = "ja-JP-Wavenet-C"
	VoiceWavenetD          = "ja-JP-Wavenet-D"
	AudioEncoding_MP3      = texttospeechpb.AudioEncoding_MP3
	AudioEncoding_LINEAR16 = texttospeechpb.AudioEncoding_LINEAR16
	AudioEncoding_OGG_OPUS = texttospeechpb.AudioEncoding_OGG_OPUS
)

type SpeechOption struct {
	LanguageCode      string
	VoiceName         string
	AudioEncoding     texttospeechpb.AudioEncoding
	AudioSpeakingRate float64
	AudioPitch        float64
}

type AudioEncoding texttospeechpb.AudioEncoding

type Speaker struct {
	client *texttospeech.Client
}

func NewSpeechClient(ctx context.Context) (*Speaker, error) {
	if speaker != nil {
		return speaker, nil
	}
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	speaker = &Speaker{client: client}
	return speaker, nil
}

func NewRequest(text string, opt *SpeechOption) *texttospeechpb.SynthesizeSpeechRequest {
	return &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: opt.LanguageCode,
			Name:         opt.VoiceName,
			SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: opt.AudioEncoding,
			SpeakingRate:  opt.AudioSpeakingRate,
			Pitch:         opt.AudioPitch,
		},
	}
}

func (s *Speaker) Run(ctx context.Context, req *texttospeechpb.SynthesizeSpeechRequest) ([]byte, error) {
	resp, err := s.client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.AudioContent, nil
}

// StreamAudioProcessor handles streaming text-to-speech processing
type StreamAudioProcessor struct {
	ctx       context.Context
	speaker   *Speaker
	option    *SpeechOption
	outputW   io.Writer
	chunkSize int
}

// NewStreamAudioProcessor creates a new streaming processor
func NewStreamAudioProcessor(ctx context.Context, speaker *Speaker, opt *SpeechOption, output io.Writer, chunkSize int) *StreamAudioProcessor {
	if chunkSize <= 0 {
		chunkSize = 1000 // Default chunk size for text
	}
	return &StreamAudioProcessor{
		ctx:       ctx,
		speaker:   speaker,
		option:    opt,
		outputW:   output,
		chunkSize: chunkSize,
	}
}

// ProcessTextStream processes text in chunks and streams the audio output
func (p *StreamAudioProcessor) ProcessTextStream(textInput io.Reader) error {
	// Read the entire file content instead of chunking it
	content, err := io.ReadAll(textInput)
	if err != nil {
		return fmt.Errorf("failed to read input: %v", err)
	}

	// Process the entire text at once to avoid malformed JSON issues
	req := NewRequest(string(content), p.option)

	audioData, err := p.speaker.Run(p.ctx, req)
	if err != nil {
		return fmt.Errorf("synthesis failed: %v", err)
	}

	// Write the audio data to the output
	if _, err := p.outputW.Write(audioData); err != nil {
		return fmt.Errorf("failed to write audio: %v", err)
	}

	return nil
}
