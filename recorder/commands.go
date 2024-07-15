package recorder

import (
	"os"
	"os/exec"
	"runtime"
)

func createRecordCommand(fullPath, format string) (*exec.Cmd, error) {
	if runtime.GOOS == "windows" {
		return exec.Command("sox", "-d", "-t", "waveaudio", fullPath), nil
	}
	return exec.Command("sox", "-d", "-t", format, fullPath), nil
}

func stopRecording(cmd *exec.Cmd) error {
	if runtime.GOOS == "windows" {
		return cmd.Process.Kill()
	}
	return cmd.Process.Signal(os.Interrupt)
}
