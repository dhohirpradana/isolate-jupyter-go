package helper

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func getYaml(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return "", err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file) // Ensure the file is closed when done

	// Create a buffer to hold the file contents
	buffer := make([]byte, 1024)

	// Read the file content
	var fileContent string
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading file:", err)
			return "", err
		}

		if n == 0 {
			break // Reached EOF
		}

		// Append the contents read to the fileContent
		fileContent += string(buffer[:n])
	}

	return fileContent, nil
}

func ReplaceString(content, old, new string) (string, error) {
	modifiedContent := strings.ReplaceAll(content, old, new)
	return modifiedContent, nil
}

func validateYamlLabel(s string) error {
	pattern := `^([a-zA-Z0-9][a-zA-Z0-9\-\._]{0,61}[a-zA-Z0-9])?$`

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %v", err)
	}

	if regex.MatchString(s) {
		return nil
	} else {
		return fmt.Errorf("invalid string: must be 63 characters or less, unless empty must begin and end with an alphanumeric character, and can contain dashes, underscores, dots, and alphanumerics between")
	}
}

func GenerateYaml(serviceName, port string) error {
	err := validateYamlLabel(serviceName)
	if err != nil {
		return errors.New("username must be 63 characters or less, unless empty must begin and end with an alphanumeric character, and can contain dashes, underscores, dots, and alphanumerics between")
	}

	filePath := "input.yaml"
	savePath := "outputs/output-" + serviceName + ".yaml"

	yaml, err := getYaml(filePath)
	if err != nil {
		fmt.Println(err.Error())
	}

	replacedString, err := ReplaceString(yaml, "SERVICE_NAME", serviceName)
	if err != nil {
		return err
	}

	replacedString, err = ReplaceString(replacedString, "SERVICE_PORT", port)
	if err != nil {
		return err
	}

	file, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	_, err = file.WriteString(replacedString)
	if err != nil {
		return err
	}

	return nil
}
