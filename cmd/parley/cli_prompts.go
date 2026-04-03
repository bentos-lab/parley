package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

var errPromptAborted = errors.New("prompt aborted")

func promptYesNo(prompt string, defaultYes bool) (bool, error) {
	value := defaultYes
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().Title(prompt).Value(&value),
		),
	).WithKeyMap(promptKeyMap())
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, errPromptAborted
		}
		return false, fmt.Errorf("interactive prompt failed: %w", err)
	}
	return value, nil
}

func promptSelect(title string, options []string, defaultIndex int) (int, error) {
	if len(options) == 0 {
		return 0, fmt.Errorf("no options provided")
	}
	if defaultIndex < 0 || defaultIndex >= len(options) {
		defaultIndex = 0
	}
	choice := options[defaultIndex]
	huhOptions := make([]huh.Option[string], 0, len(options))
	for _, option := range options {
		huhOptions = append(huhOptions, huh.NewOption(option, option))
	}
	selectField := huh.NewSelect[string]().Title(title).Options(huhOptions...).Value(&choice)
	form := huh.NewForm(huh.NewGroup(selectField)).WithKeyMap(promptKeyMap())
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return 0, errPromptAborted
		}
		return 0, fmt.Errorf("interactive prompt failed: %w", err)
	}
	for i, option := range options {
		if option == choice {
			return i, nil
		}
	}
	return defaultIndex, nil
}

func promptRequired(prompt string) (string, error) {
	if runtime.GOOS == "windows" {
		// fallback
		fmt.Print(prompt + " ")
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			return "", fmt.Errorf("value is required")
		}
		return text, nil
	}

	// dùng huh cho non-windows
	value := ""
	input := huh.NewInput().
		Title(prompt).
		Value(&value).
		Validate(func(text string) error {
			if strings.TrimSpace(text) == "" {
				return fmt.Errorf("value is required")
			}
			return nil
		})

	form := huh.NewForm(huh.NewGroup(input)).WithKeyMap(promptKeyMap())
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", errPromptAborted
		}
		return "", fmt.Errorf("interactive prompt failed: %w", err)
	}

	return strings.TrimSpace(value), nil
}

func promptKeyMap() *huh.KeyMap {
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "quit"))
	return keymap
}
