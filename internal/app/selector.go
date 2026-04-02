package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LiukerSun/cc-cli/internal/config"
	"golang.org/x/term"
)

var stdinIsInteractive = isInteractiveInput
var arrowSelectorEnabled = supportsArrowSelector
var makeRawSelectorInput = makeRawSelectorMode

func selectRunProfile(stdin io.Reader, stdout io.Writer, cfg config.File) (string, error) {
	if arrowSelectorEnabled(stdin, stdout) {
		return selectRunProfileWithArrows(stdin, stdout, cfg)
	}
	return selectRunProfileWithNumbers(stdin, stdout, cfg)
}

func selectRunProfileWithNumbers(stdin io.Reader, stdout io.Writer, cfg config.File) (string, error) {
	reader := bufio.NewReader(stdin)

	fmt.Fprintln(stdout, "ccc run")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Select a profile:")
	for i, profile := range cfg.Profiles {
		current := ""
		if profile.ID == cfg.CurrentProfile {
			current = " [current]"
		}
		fmt.Fprintf(stdout, "  %2d) %s%s\n", i+1, profile.Name, current)
		fmt.Fprintf(stdout, "      %s | %s | %s\n", profile.Command, profile.Provider, profile.Model)
	}

	defaultChoice := "1"
	if cfg.CurrentProfile != "" {
		for i, profile := range cfg.Profiles {
			if profile.ID == cfg.CurrentProfile {
				defaultChoice = fmt.Sprintf("%d", i+1)
				break
			}
		}
	}

	for {
		value, err := promptWithDefault(reader, stdout, fmt.Sprintf("Choice [1-%d]", len(cfg.Profiles)), defaultChoice)
		if err != nil {
			return "", err
		}
		var choice int
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &choice); err == nil && choice >= 1 && choice <= len(cfg.Profiles) {
			return cfg.Profiles[choice-1].ID, nil
		}
		fmt.Fprintf(stdout, "Please enter a number between 1 and %d.\n", len(cfg.Profiles))
	}
}

func selectRunProfileWithArrows(stdin io.Reader, stdout io.Writer, cfg config.File) (string, error) {
	restore, err := makeRawSelectorInput(stdin)
	if err != nil {
		return "", err
	}
	if restore != nil {
		defer restore()
	}

	reader := bufio.NewReader(stdin)
	selected := 0
	if cfg.CurrentProfile != "" {
		for i, profile := range cfg.Profiles {
			if profile.ID == cfg.CurrentProfile {
				selected = i
				break
			}
		}
	}

	printArrowSelectorList(stdout, cfg)
	lastLineLength := renderArrowSelectorStatus(stdout, cfg, selected, 0)
	for {
		key, err := readSelectorKey(reader)
		if err != nil {
			return "", err
		}

		switch key {
		case selectorKeyUp:
			if selected == 0 {
				selected = len(cfg.Profiles) - 1
			} else {
				selected--
			}
			lastLineLength = renderArrowSelectorStatus(stdout, cfg, selected, lastLineLength)
		case selectorKeyDown:
			if selected == len(cfg.Profiles)-1 {
				selected = 0
			} else {
				selected++
			}
			lastLineLength = renderArrowSelectorStatus(stdout, cfg, selected, lastLineLength)
		case selectorKeyEnter:
			fmt.Fprint(stdout, "\r\n")
			return cfg.Profiles[selected].ID, nil
		case selectorKeyCancel:
			fmt.Fprint(stdout, "\r\n")
			return "", fmt.Errorf("selection cancelled")
		}
	}
}

type selectorKey int

const (
	selectorKeyUnknown selectorKey = iota
	selectorKeyUp
	selectorKeyDown
	selectorKeyEnter
	selectorKeyCancel
)

func printArrowSelectorList(stdout io.Writer, cfg config.File) {
	writeSelectorLine(stdout, "ccc")
	writeSelectorLine(stdout, "")
	writeSelectorLine(stdout, "Use Up/Down to choose a profile, Enter to run, q to quit.")
	writeSelectorLine(stdout, "")
	for i, profile := range cfg.Profiles {
		current := ""
		if profile.ID == cfg.CurrentProfile {
			current = " [current]"
		}
		writeSelectorLine(stdout, fmt.Sprintf("  %2d) %s%s", i+1, profile.Name, current))
		writeSelectorLine(stdout, fmt.Sprintf("      %s | %s | %s", profile.Command, profile.Provider, profile.Model))
	}
}

func writeSelectorLine(stdout io.Writer, line string) {
	_, _ = io.WriteString(stdout, line+"\r\n")
}

func renderArrowSelectorStatus(stdout io.Writer, cfg config.File, selected, previousLength int) int {
	profile := cfg.Profiles[selected]
	current := ""
	if profile.ID == cfg.CurrentProfile {
		current = " [current]"
	}
	line := fmt.Sprintf("Selected: %d) %s%s", selected+1, profile.Name, current)
	padding := ""
	if previousLength > len(line) {
		padding = strings.Repeat(" ", previousLength-len(line))
	}
	fmt.Fprintf(stdout, "\r%s%s", line, padding)
	return len(line)
}

func readSelectorKey(reader *bufio.Reader) (selectorKey, error) {
	first, err := reader.ReadByte()
	if err != nil {
		return selectorKeyUnknown, err
	}

	switch first {
	case '\r', '\n':
		return selectorKeyEnter, nil
	case 'q', 'Q', 3:
		return selectorKeyCancel, nil
	case 'k', 'K':
		return selectorKeyUp, nil
	case 'j', 'J':
		return selectorKeyDown, nil
	case 0x1b:
		second, err := reader.ReadByte()
		if err != nil {
			return selectorKeyUnknown, err
		}
		if second != '[' {
			return selectorKeyUnknown, nil
		}
		third, err := reader.ReadByte()
		if err != nil {
			return selectorKeyUnknown, err
		}
		switch third {
		case 'A':
			return selectorKeyUp, nil
		case 'B':
			return selectorKeyDown, nil
		default:
			return selectorKeyUnknown, nil
		}
	case 0x00, 0xe0:
		second, err := reader.ReadByte()
		if err != nil {
			return selectorKeyUnknown, err
		}
		switch second {
		case 'H':
			return selectorKeyUp, nil
		case 'P':
			return selectorKeyDown, nil
		default:
			return selectorKeyUnknown, nil
		}
	default:
		return selectorKeyUnknown, nil
	}
}

func isInteractiveInput(stdin io.Reader) bool {
	file, ok := stdin.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func supportsArrowSelector(stdin io.Reader, stdout io.Writer) bool {
	inputFile, ok := stdin.(*os.File)
	if !ok || !term.IsTerminal(int(inputFile.Fd())) {
		return false
	}
	outputFile, ok := stdout.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(outputFile.Fd()))
}

func makeRawSelectorMode(stdin io.Reader) (func(), error) {
	inputFile, ok := stdin.(*os.File)
	if !ok {
		return nil, nil
	}
	state, err := term.MakeRaw(int(inputFile.Fd()))
	if err != nil {
		return nil, err
	}
	return func() {
		_ = term.Restore(int(inputFile.Fd()), state)
	}, nil
}
