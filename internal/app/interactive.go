package app

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"golang.org/x/term"

	"github.com/LiukerSun/cc-cli/internal/util"
)

type fileDescriptorInput interface {
	Fd() uintptr
}

var secretInputSupported = supportsHiddenSecretInput
var readSecretInput = readHiddenSecretInput

func promptRequired(reader *bufio.Reader, stdout io.Writer, label string) (string, error) {
	for {
		value, err := promptWithDefault(reader, stdout, label, "")
		if err != nil {
			return "", err
		}
		value = strings.TrimSpace(value)
		if value != "" {
			return value, nil
		}
		fmt.Fprintf(stdout, "%s is required.\n", label)
	}
}

func promptWithDefault(reader *bufio.Reader, stdout io.Writer, label, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Fprintf(stdout, "%s [%s]: ", label, defaultValue)
	} else {
		fmt.Fprintf(stdout, "%s: ", label)
	}

	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	value := strings.TrimSpace(line)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func promptSecretRequired(stdin io.Reader, reader *bufio.Reader, stdout io.Writer, label string) (string, error) {
	for {
		value, err := promptSecret(stdin, reader, stdout, label)
		if err != nil {
			return "", err
		}
		value = strings.TrimSpace(value)
		if value != "" {
			return value, nil
		}
		fmt.Fprintf(stdout, "%s is required.\n", label)
	}
}

func promptSecret(stdin io.Reader, reader *bufio.Reader, stdout io.Writer, label string) (string, error) {
	if secretInputSupported(stdin) {
		return readSecretInput(stdin, stdout, label)
	}
	return promptWithDefault(reader, stdout, label, "")
}

func promptChoice(reader *bufio.Reader, stdout io.Writer, label string, max int) (int, error) {
	for {
		value, err := promptRequired(reader, stdout, fmt.Sprintf("%s [1-%d]", label, max))
		if err != nil {
			return 0, err
		}
		var choice int
		if _, err := fmt.Sscanf(value, "%d", &choice); err == nil && choice >= 1 && choice <= max {
			return choice, nil
		}
		fmt.Fprintf(stdout, "Please enter a number between 1 and %d.\n", max)
	}
}

func promptModelChoice(reader *bufio.Reader, stdout io.Writer, kind string, models []string, defaultValue string, allowCustom bool) (string, error) {
	choices := util.UniqueStrings(models)
	if allowCustom {
		choices = append(choices, "Custom model ID")
	}

	fmt.Fprintf(stdout, "Available %s models:\n", kind)
	for i, model := range choices {
		fmt.Fprintf(stdout, "  %2d) %s\n", i+1, model)
	}

	if defaultValue != "" {
		if idx := indexOf(choices, defaultValue); idx >= 0 {
			value, err := promptWithDefault(reader, stdout, fmt.Sprintf("Select %s model [1-%d]", kind, len(choices)), fmt.Sprintf("%d", idx+1))
			if err != nil {
				return "", err
			}
			return resolveModelChoice(reader, stdout, choices, value)
		}
		if allowCustom {
			value, err := promptWithDefault(reader, stdout, fmt.Sprintf("Select %s model [1-%d]", kind, len(choices)), fmt.Sprintf("%d", len(choices)))
			if err != nil {
				return "", err
			}
			model, err := resolveModelChoice(reader, stdout, choices, value)
			if err != nil {
				return "", err
			}
			if model == "Custom model ID" {
				return promptRequired(reader, stdout, "Custom model ID")
			}
			return model, nil
		}
	}

	for {
		value, err := promptRequired(reader, stdout, fmt.Sprintf("Select %s model [1-%d]", kind, len(choices)))
		if err != nil {
			return "", err
		}
		model, err := resolveModelChoice(reader, stdout, choices, value)
		if err != nil {
			fmt.Fprintln(stdout, err.Error())
			continue
		}
		if model == "Custom model ID" {
			return promptRequired(reader, stdout, "Custom model ID")
		}
		return model, nil
	}
}

func resolveModelChoice(reader *bufio.Reader, stdout io.Writer, choices []string, value string) (string, error) {
	var choice int
	if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &choice); err != nil || choice < 1 || choice > len(choices) {
		return "", fmt.Errorf("please enter a number between 1 and %d", len(choices))
	}
	model := choices[choice-1]
	if model == "Custom model ID" {
		customModel, err := promptRequired(reader, stdout, "Custom model ID")
		if err != nil {
			return "", err
		}
		return customModel, nil
	}
	return model, nil
}

func supportsHiddenSecretInput(stdin io.Reader) bool {
	input, ok := stdin.(fileDescriptorInput)
	if !ok {
		return false
	}
	return term.IsTerminal(int(input.Fd()))
}

func readHiddenSecretInput(stdin io.Reader, stdout io.Writer, label string) (string, error) {
	input, ok := stdin.(fileDescriptorInput)
	if !ok {
		return "", fmt.Errorf("hidden input is not available")
	}

	fmt.Fprintf(stdout, "%s: ", label)
	line, err := term.ReadPassword(int(input.Fd()))
	fmt.Fprintln(stdout)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(line)), nil
}
