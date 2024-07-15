package main

import (
	"context"
	"flag"
	"fmt"

	// "io"
	"os"
	// "syscall"
	"log"
	"path/filepath"
	"time"

	// "github.com/manifoldco/promptui"
	"github.com/r4h4/article-helper/editor"
	"github.com/r4h4/article-helper/recorder"

	// "github.com/r4h4/article-helper/downloader"
	"github.com/r4h4/article-helper/whisper"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	outputFile := flag.String("o", "", "Output file name (default: current timestamp)")
	// model := flag.String("m", "", "Model to use for transcription")
	flag.Parse()

	// Create output folder
	timestamp := time.Now().Format("20060102_150405")
	outFolder := fmt.Sprintf("./recordings/%s", timestamp)
	if _, err := os.Stat(outFolder); os.IsNotExist(err) {
		os.MkdirAll(outFolder, 0755)
	}

	if *outputFile == "" {
		*outputFile = fmt.Sprintf("recording_%s.wav", timestamp)
	}

	audioOptions := recorder.ConfigurableOptions{
		RecordingsDir: outFolder,
		AudioFormat:   "wav",
	}
	// Start the audio recording
	if err := recorder.RecordAudio(context.TODO(), *outputFile, audioOptions); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the OPENAI_API_KEY environment variable")
		return
	}
	config := whisper.Config{
		APIEndpoint: "https://api.openai.com/v1/audio/transcriptions",
		APIKey:      apiKey,
	}

	client := whisper.NewClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// filePath := "path/to/your/audio/file.mp3"
	transcription, err := client.TranscribeAudio(
		ctx, filepath.Join(audioOptions.RecordingsDir, *outputFile),
	)
	if err != nil {
		log.Fatalf("Error transcribing audio: %v", err)
	}

	// Save the transcription to a file
	transcriptionFile := fmt.Sprintf("transcription_%s.txt", timestamp)
	if err := os.WriteFile(filepath.Join(outFolder, transcriptionFile), []byte(transcription), 0644); err != nil {
		log.Fatalf("Error saving transcription to file: %v", err)
	}

	log.Println("Starting Editor...")
	editor := editor.NewAIEditor(apiKey, "")
	result, err := editor.EditAndSummarize(transcription)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Save the cleaned transcription and summary to one file each
	cleanedTranscriptionFile := fmt.Sprintf("cleaned_transcription_%s.txt", timestamp)
	summaryFile := fmt.Sprintf("summary_%s.txt", timestamp)

	if err := os.WriteFile(filepath.Join(outFolder, cleanedTranscriptionFile), []byte(result.CleanedTranscription), 0644); err != nil {
		log.Fatalf("Error saving cleaned transcription to file: %v", err)
	}

	if err := os.WriteFile(filepath.Join(outFolder, summaryFile), []byte(result.Summary), 0644); err != nil {
		log.Fatalf("Error saving summary to file: %v", err)
	}

	fmt.Printf("Cleaned Transcription:\n%s\n\n", result.CleanedTranscription)
	fmt.Printf("Summary:\n%s\n", result.Summary)

	headline, err := editor.CreateHeadline(result.Summary)
	if err != nil {
		fmt.Printf("Error creating headline: %v\n", err)
		return
	}
	// Rename the folder to the headline
	newFolderName := fmt.Sprintf("./recordings/%s_%s", timestamp, headline)
	if err := os.Rename(outFolder, newFolderName); err != nil {
		log.Fatalf("Error renaming folder: %v", err)
	}
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
