package recorder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/eiannone/keyboard"
)

// ConfigurableOptions allows for easier extension and configuration
type ConfigurableOptions struct {
	RecordingsDir string
	AudioFormat   string
}

// RecordAudio records audio in the terminal and saves the output to a file
func RecordAudio(ctx context.Context, outputFile string, opts ConfigurableOptions) error {
	if err := ensureDir(opts.RecordingsDir); err != nil {
		return fmt.Errorf("failed to create recordings directory: %w", err)
	}

	fullPath := filepath.Join(opts.RecordingsDir, outputFile)

	cmd, err := createRecordCommand(fullPath, opts.AudioFormat)
	if err != nil {
		return fmt.Errorf("failed to create record command: %w", err)
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	fmt.Println("Recording started. Press ESC to stop...")

	// Setup signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Setup keyboard listening
	keyChan := make(chan keyboard.Key, 1)
	go listenForEscKey(ctx, keyChan)

	// Wait for stop signal
	select {
	case <-signalChan:
		fmt.Println("\nReceived interrupt signal. Stopping recording...")
	case key := <-keyChan:
		if key == keyboard.KeyEsc {
			fmt.Println("ESC pressed. Stopping recording...")
		}
	case <-ctx.Done():
		fmt.Println("Context cancelled. Stopping recording...")
	}

	// Stop the recording
	if err := stopRecording(cmd); err != nil {
		fmt.Printf("Error stopping recording: %v\n", err)
	}

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// On Unix-like systems, exit status 1 might be normal for interrupted processes
			if exitErr.ExitCode() != 1 || runtime.GOOS == "windows" {
				return fmt.Errorf("error during recording: %w", err)
			}
		} else {
			return fmt.Errorf("error during recording: %w", err)
		}
	}

	fmt.Printf("Audio recorded successfully: %s\n", fullPath)
	return nil
}
