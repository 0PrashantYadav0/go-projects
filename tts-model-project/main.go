package main

import (
	"os"
	"tts-model-project/cmd"
)

func main() {
	cli := &cmd.CLI{ErrStream: os.Stderr}
	os.Exit(cli.Run(os.Args))
}
