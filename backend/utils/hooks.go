package utils

import (
	"fmt"
	"os/exec"

	"github.com/google/shlex"
)

// ExecuteCommandBg runs the given command in the background without waiting for it to finish.
func ExecuteCommandBg(hookCommand string) {
	if hookCommand == "" {
		return
	}

	// Split the command into name and args
	parts, err := shlex.Split(hookCommand)
	if err != nil || len(parts) == 0 {
		return
	}

	cmd := exec.Command(parts[0], parts[1:]...)

	// Run the command in the background
	err = cmd.Start()
	if err != nil {
		return
	}

	// wait for the command to finish in a separate goroutine
	go func() {
		err := cmd.Wait()

		if err != nil {
			fmt.Println("Hook command failed:", err)
			return
		}
	}()
}
