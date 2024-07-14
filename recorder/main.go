package recorder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/eiannone/keyboard"
)

func ensureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func RecordAudio(outputFile string) error {
	recordingsDir := "./recordings"
	if err := ensureDir(recordingsDir); err != nil {
		return fmt.Errorf("failed to create recordings directory: %w", err)
	}

	fullPath := filepath.Join(recordingsDir, outputFile)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("sox", "-d", "-t", "waveaudio", fullPath)
	} else {
		cmd = exec.Command("sox", "-d", "-t", "wav", fullPath)
	}

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	fmt.Println("Recording started. Press ESC to stop...")

	if err := keyboard.Open(); err != nil {
		return fmt.Errorf("failed to open keyboard: %w", err)
	}

	// Use a goroutine to listen for the ESC key
	stopChan := make(chan struct{})
	go func() {
		defer keyboard.Close()
		for {
			_, key, err := keyboard.GetKey()
			if err != nil {
				fmt.Println("Error reading keyboard:", err)
				return
			}
			if key == keyboard.KeyEsc {
				close(stopChan)
				return
			}
		}
	}()

	// Wait for either the recording to finish or the stop signal
	<-stopChan

	// Stop the recording
	if runtime.GOOS == "windows" {
		cmd.Process.Kill()
	} else {
		cmd.Process.Signal(os.Interrupt)
	}

	err = cmd.Wait()
	if err != nil && runtime.GOOS != "windows" {
		return fmt.Errorf("error during recording: %w", err)
	}

	fmt.Printf("Audio recorded successfully: %s\n", fullPath)
	return nil
}
