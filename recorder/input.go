package recorder

import (
	"context"
	"fmt"

	"github.com/eiannone/keyboard"
)

func listenForEscKey(ctx context.Context, keyChan chan<- keyboard.Key) {
	if err := keyboard.Open(); err != nil {
		fmt.Println("Failed to open keyboard:", err)
		return
	}
	defer keyboard.Close()

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Println("Error reading keyboard:", err)
			return
		}
		if key == keyboard.KeyEsc || char == 'q' || char == 'Q' {
			keyChan <- key
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			// Continue listening
		}
	}
}
