package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("loading .env file: %w", err)
	}

	outputFile := flag.String("o", "", "Output file name (default: current timestamp)")
	flag.Parse()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	ctx := context.Background()
	timestamp := time.Now().Format("20060102_150405")

	pipeline := []Step{
		&RecordStep{OutputFile: outputFile},
		&TranscribeStep{APIKey: apiKey},
		&EditStep{APIKey: apiKey},
		&SaveStep{},
		&HeadlineStep{APIKey: apiKey},
	}

	state := &State{
		Timestamp: timestamp,
		OutFolder: fmt.Sprintf("./recordings/%s", timestamp),
	}

	for _, step := range pipeline {
		if err := step.Execute(ctx, state); err != nil {
			return fmt.Errorf("%T: %w", step, err)
		}
	}

	return nil
}

// Artifacts of local mode
// Get available models
// models := downloader.GetModels()

// // If a model is provided, check if it is valid
// if *model != "" {
// 	validModel := false
// 	for _, m := range models {
// 		if *model == m {
// 			validModel = true
// 			break
// 		}
// 	}

// 	if !validModel {
// 		fmt.Fprintf(os.Stderr, "Invalid model: %s\n", *model)
// 		os.Exit(1)
// 	}
// } else {
// 	// Prompt user to select a model
// 	prompt := promptui.Select{
// 		Label: "Select a transcriber model",
// 		Items: models,
// 	}

// 	_, selectedModel, err := prompt.Run()

// 	if err != nil {
// 		fmt.Printf("Prompt failed %v\n", err)
// 		return
// 	}

// 	// Check if the model is already downloaded
// 	if !downloader.IsModelDownloaded(selectedModel, "./models") {
// 		prompt := promptui.Prompt{
// 			Label:     "The model is not downloaded. Do you want to download it now?",
// 			IsConfirm: true,
// 		}

// 		result, err := prompt.Run()

// 		if err != nil {
// 			fmt.Printf("Aborted %v\n", err)
// 			return
// 		}
// 		fmt.Printf("Result: %v\n", result)

// 		// Create context which quits on SIGINT or SIGQUIT
// 		ctx := downloader.ContextForSignal(os.Interrupt, syscall.SIGQUIT)

// 		progress := os.Stdout

// 		// Download models - exit on error or interrupt
// 		url, err := downloader.URLForModel(selectedModel)
// 		if err != nil {
// 			fmt.Fprintln(os.Stderr, "Error:", err)
// 			os.Exit(-1)
// 		} else if path, err := downloader.Download(ctx, progress, url, "./models"); err == nil || err == io.EOF {
// 			fmt.Println("Model downloaded successfully")
// 		} else if err == context.Canceled {
// 			os.Remove(path)
// 			fmt.Fprintln(progress, "\nInterrupted")
// 			os.Exit(-1)
// 		} else if err == context.DeadlineExceeded {
// 			os.Remove(path)
// 			fmt.Fprintln(progress, "Timeout downloading model")
// 			os.Exit(-1)
// 		} else {
// 			os.Remove(path)
// 			fmt.Fprintln(os.Stderr, "Error:", err)
// 			os.Exit(-1)
// 		}
// 	}
// }
