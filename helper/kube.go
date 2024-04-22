package helper

import (
	"fmt"
	"os/exec"
	"strings"
)

func UnusedPort() (string, error) {
	scriptPath := "unused_port.sh"
	cmd := exec.Command("/bin/bash", scriptPath)

	outputChan := make(chan string)
	errChan := make(chan error)

	go func() {
		output, err := cmd.Output()
		if err != nil {
			errChan <- err
			return
		}
		unusedPort := strings.TrimSpace(string(output))
		outputChan <- unusedPort
	}()

	select {
	case unusedPort := <-outputChan:
		return unusedPort, nil
	case err := <-errChan:
		fmt.Printf("Error executing script: %v\n", err)
		return "", err
	}
}
