package helper

import (
	"fmt"
	"os/exec"
	"strings"
)

func UnusedPort() (string, error) {
	scriptPath := "unused_port.sh"
	cmd := exec.Command("/bin/bash", scriptPath)

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error executing script: %v\n", err)
		return "", err
	}

	unusedPort := string(output)
	unusedPort = strings.TrimSpace(unusedPort)

	return unusedPort, nil
}
