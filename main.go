package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/r4h4/article-helper/recorder"
	"github.com/r4h4/article-helper/transcriber"
)

func main() {
	outputFile := flag.String("o", "", "Output file name (default: current timestamp)")
	model := flag.String("m", "ggml-medium.en", "Model to use for transcription")
	flag.Parse()

	if *outputFile == "" {
		*outputFile = fmt.Sprintf("recording_%s.wav", time.Now().Format("20060102_150405"))
	}

	if err := recorder.RecordAudio(*outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	transcriber.Process(*model, *outputFile, nil)

	fmt.Println("Recording stopped. You can add more functionality here.")
	// Add your additional functionality here
}
