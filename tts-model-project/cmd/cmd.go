package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"tts-model-project/speech"

	"flag"
)

const (
	ExitCodeOK              = 0
	ExitCodeParseFlagsError = 1
	ExitCodeValidateError   = 2
	ExitCodeInternalError   = 3
	ExitCodeOutputFileError = 4
)

type CLI struct {
	ErrStream io.Writer
}

func (cli *CLI) Run(args []string) int {
	flags := flag.NewFlagSet("google-text-to-speech", flag.ContinueOnError)
	var (
		text, voice, out, inputFile string
		rate, pitch                 float64
		streaming                   bool
		chunkSize                   int
	)
	flags.StringVar(&text, "text", "", "text to speech")
	flags.StringVar(&voice, "voice", "stand-a", "speaker's voice name")
	flags.Float64Var(&rate, "rate", 1.00, "speech rate(0.25 ~ 4.0)")
	flags.Float64Var(&pitch, "pitch", 0.00, "speaking pitch(-20.0 ~ 20.0)")
	flags.StringVar(&out, "o", "", "output audio file(support format of the audio: LINEAR16 , MP3)")
	flags.StringVar(&inputFile, "input-file", "", "input text file for streaming (use with --streaming)")
	flags.BoolVar(&streaming, "streaming", false, "use streaming mode")
	flags.IntVar(&chunkSize, "chunk-size", 1000, "text chunk size for streaming")

	if err := flags.Parse(args[1:]); err != nil {
		fmt.Fprint(cli.ErrStream, err)
		return ExitCodeParseFlagsError
	}

	// Modified validation logic to handle streaming mode
	var opt *speech.SpeechOption
	var err error

	if streaming {
		// For streaming, we pass a dummy text to bypass validation
		// The actual text will come from the input file
		opt, err = makeSpeechOpt("dummy-text-for-streaming", voice, out, rate, pitch)
	} else {
		opt, err = makeSpeechOpt(text, voice, out, rate, pitch)
	}

	if err != nil {
		fmt.Fprint(cli.ErrStream, err)
		return ExitCodeValidateError
	}

	ctx := context.Background()
	speaker, err := speech.NewSpeechClient(ctx)
	if err != nil {
		fmt.Fprint(cli.ErrStream, err)
		return ExitCodeInternalError
	}

	if streaming {
		if inputFile == "" {
			fmt.Fprintf(cli.ErrStream, "Streaming mode requires an input file (--input-file)\n")
			return ExitCodeParseFlagsError
		}
		return cli.handleStreaming(ctx, speaker, opt, inputFile, out, chunkSize)
	} else {
		if text == "" {
			fmt.Fprintf(cli.ErrStream, "Non-streaming mode requires text input (-text)\n")
			return ExitCodeValidateError
		}
		return cli.handleNonStreaming(ctx, speaker, opt, text, out)
	}
}

func (cli *CLI) handleNonStreaming(ctx context.Context, speaker *speech.Speaker, opt *speech.SpeechOption, text, out string) int {
	req := speech.NewRequest(text, opt)
	audioData, err := speaker.Run(ctx, req)
	if err != nil {
		fmt.Fprint(cli.ErrStream, err)
		return ExitCodeInternalError
	}

	if err = os.WriteFile(out, audioData, 0644); err != nil {
		fmt.Fprint(cli.ErrStream, err)
		return ExitCodeOutputFileError
	}

	fmt.Printf("Audio file created successfully at: %s\n", out)
	return ExitCodeOK
}

func (cli *CLI) handleStreaming(ctx context.Context, speaker *speech.Speaker, opt *speech.SpeechOption, inputFile, outFile string, chunkSize int) int {
	var textInput io.Reader

	if inputFile != "" {
		file, err := os.Open(inputFile)
		if err != nil {
			fmt.Fprintf(cli.ErrStream, "Failed to open input file: %v\n", err)
			return ExitCodeParseFlagsError
		}
		defer file.Close()
		textInput = file
	} else {
		fmt.Fprintf(cli.ErrStream, "Streaming mode requires an input file\n")
		return ExitCodeParseFlagsError
	}

	outF, err := os.Create(outFile)
	if err != nil {
		fmt.Fprintf(cli.ErrStream, "Failed to create output file: %v\n", err)
		return ExitCodeOutputFileError
	}
	defer outF.Close()

	processor := speech.NewStreamAudioProcessor(ctx, speaker, opt, outF, chunkSize)
	if err := processor.ProcessTextStream(textInput); err != nil {
		fmt.Fprintf(cli.ErrStream, "Error during streaming processing: %v\n", err)
		return ExitCodeInternalError
	}

	fmt.Printf("Streaming audio created successfully at: %s\n", outFile)
	return ExitCodeOK
}

func makeSpeechOpt(text, voice, out string, rate, pitch float64) (*speech.SpeechOption, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text")
	}

	var voiceName string
	switch v := strings.ToLower(voice); v {
	case "stand-a":
		voiceName = speech.VoiceStandardA
	case "stand-b":
		voiceName = speech.VoiceStandardB
	case "stand-c":
		voiceName = speech.VoiceStandardC
	case "stand-d":
		voiceName = speech.VoiceStandardD
	case "wavenet-a":
		voiceName = speech.VoiceWavenetA
	case "wavenet-b":
		voiceName = speech.VoiceWavenetB
	case "wavenet-c":
		voiceName = speech.VoiceWavenetC
	case "wavenet-d":
		voiceName = speech.VoiceWavenetD
	default:
		return nil, fmt.Errorf("unknown voiceName: %v", v)
	}

	if 0.25 > rate || rate > 4.0 {
		return nil, fmt.Errorf("valid speaking_rate is between 0.25 and 4.0 (rate: %g)", rate)
	}

	if -20.00 > pitch || pitch > 20.00 {
		return nil, fmt.Errorf("valid pitch is between -20.0 and 20.0 (pitch: %g)", pitch)
	}

	switch ext := strings.ToLower(filepath.Ext(out)); ext {
	case ".wav":
		return &speech.SpeechOption{
			LanguageCode:      "ja-JP",
			VoiceName:         voiceName,
			AudioEncoding:     speech.AudioEncoding_LINEAR16,
			AudioSpeakingRate: rate,
			AudioPitch:        pitch,
		}, nil
	case ".mp3":
		return &speech.SpeechOption{
			LanguageCode:      "ja-JP",
			VoiceName:         voiceName,
			AudioEncoding:     speech.AudioEncoding_MP3,
			AudioSpeakingRate: rate,
			AudioPitch:        pitch,
		}, nil
	default:
		return nil, fmt.Errorf("unknown extension (out:%s)", out)
	}
}

// cb18485@canara
