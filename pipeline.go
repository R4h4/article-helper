package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/r4h4/article-helper/editor"
	"github.com/r4h4/article-helper/recorder"
	"github.com/r4h4/article-helper/whisper"
)

type State struct {
	Timestamp            string
	OutFolder            string
	OutputFile           string
	Transcription        string
	CleanedTranscription string
	Summary              string
	Headline             string
}

type Step interface {
	Execute(ctx context.Context, state *State) error
}

type RecordStep struct {
	OutputFile *string
}

type TranscribeStep struct {
	APIKey string
}

type EditStep struct {
	APIKey string
}

type SaveStep struct{}

type HeadlineStep struct {
	APIKey string
}

func (s *RecordStep) Execute(ctx context.Context, state *State) error {
	if err := os.MkdirAll(state.OutFolder, 0755); err != nil {
		return fmt.Errorf("creating output folder: %w", err)
	}

	if *s.OutputFile == "" {
		state.OutputFile = fmt.Sprintf("recording_%s.wav", state.Timestamp)
	} else {
		state.OutputFile = *s.OutputFile
	}

	audioOptions := recorder.ConfigurableOptions{
		RecordingsDir: state.OutFolder,
		AudioFormat:   "wav",
	}

	if err := recorder.RecordAudio(ctx, state.OutputFile, audioOptions); err != nil {
		return fmt.Errorf("recording audio: %w", err)
	}

	return nil
}

func (s *TranscribeStep) Execute(ctx context.Context, state *State) error {
	config := whisper.Config{
		APIEndpoint: "https://api.openai.com/v1/audio/transcriptions",
		APIKey:      s.APIKey,
	}
	client := whisper.NewClient(config)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	filePath := filepath.Join(state.OutFolder, state.OutputFile)
	transcription, err := client.TranscribeAudio(ctx, filePath)
	if err != nil {
		return fmt.Errorf("transcribing audio: %w", err)
	}

	state.Transcription = transcription
	return nil
}

func (s *EditStep) Execute(ctx context.Context, state *State) error {
	ed := editor.NewAIEditor(s.APIKey, "")
	result, err := ed.EditAndSummarize(state.Transcription)
	if err != nil {
		return fmt.Errorf("editing and summarizing: %w", err)
	}
	state.CleanedTranscription = result.CleanedTranscription
	state.Summary = result.Summary
	return nil
}

func (s *SaveStep) Execute(ctx context.Context, state *State) error {
	files := map[string]string{
		fmt.Sprintf("transcription_%s.txt", state.Timestamp):         state.Transcription,
		fmt.Sprintf("cleaned_transcription_%s.txt", state.Timestamp): state.CleanedTranscription,
		fmt.Sprintf("summary_%s.txt", state.Timestamp):               state.Summary,
	}

	for fileName, content := range files {
		filePath := filepath.Join(state.OutFolder, fileName)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("saving %s: %w", fileName, err)
		}
	}

	fmt.Printf("Cleaned Transcription:\n%s\n\n", state.CleanedTranscription)
	fmt.Printf("Summary:\n%s\n", state.Summary)

	return nil
}

func (s *HeadlineStep) Execute(ctx context.Context, state *State) error {
	ed := editor.NewAIEditor(s.APIKey, "")
	result, err := ed.CreateHeadline(state.Summary)
	if err != nil {
		return fmt.Errorf("creating headline: %w", err)
	}
	state.Headline = result.Headline

	newFolderName := filepath.Join("./recordings", fmt.Sprintf("%s_%s", state.Timestamp, state.Headline))
	if err := os.Rename(state.OutFolder, newFolderName); err != nil {
		return fmt.Errorf("renaming folder: %w", err)
	}

	return nil
}
