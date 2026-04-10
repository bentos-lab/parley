package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/adapter/outbound/tts/native"
	"github.com/bentos-lab/parley/config"
	"github.com/bentos-lab/parley/shared/install"
)

const providerCustom = "custom"

func newLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Guided login for providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Help(); err != nil {
				return err
			}
			return fmt.Errorf("missing login target")
		},
	}
	cmd.AddCommand(newLoginLLMCommand())
	cmd.AddCommand(newLoginTTSCommand())
	return cmd
}

func newLoginLLMCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "llm",
		Short: "Configure the LLM provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			providers := []string{"openai", "anthropic", "gemini", providerCustom}
			labels := []string{"OpenAI", "Anthropic", "Gemini", "OpenAI-Compatible (Custom)"}
			index, err := promptSelect("Select provider:", labels, 0)
			if err != nil {
				if errors.Is(err, errPromptAborted) {
					return nil
				}
				return err
			}
			provider := providers[index]
			var handlerErr error
			switch provider {
			case "openai", "anthropic", "gemini", providerCustom:
				handlerErr = handleExistingProvider(provider)
			default:
				return fmt.Errorf("unsupported provider %q", provider)
			}
			if handlerErr != nil {
				if errors.Is(handlerErr, errPromptAborted) {
					return nil
				}
				return handlerErr
			}
			return nil
		},
	}
}

func handleExistingProvider(provider string) error {
	apiKey, err := promptRequired("Enter API key:")
	if err != nil {
		return err
	}
	model, err := promptModelSelection(provider)
	if err != nil {
		return err
	}
	cfgMap, err := config.ReadFileMap()
	if err != nil {
		return err
	}
	displayBaseURL := ""
	switch provider {
	case "openai":
		baseURL := "https://api.openai.com/v1"
		displayBaseURL = baseURL
		config.SetNestedValue(cfgMap, []string{"llm"}, "provider", "openai")
		config.SetNestedValue(cfgMap, []string{"llm", "openai"}, "base_url", baseURL)
		config.SetNestedValue(cfgMap, []string{"llm", "openai"}, "api_key", apiKey)
		config.SetNestedValue(cfgMap, []string{"llm", "openai"}, "model", model)
	case "anthropic":
		config.SetNestedValue(cfgMap, []string{"llm"}, "provider", "anthropic")
		config.SetNestedValue(cfgMap, []string{"llm", "anthropic"}, "api_key", apiKey)
		config.SetNestedValue(cfgMap, []string{"llm", "anthropic"}, "model", model)
	case "gemini":
		config.SetNestedValue(cfgMap, []string{"llm"}, "provider", "gemini")
		config.SetNestedValue(cfgMap, []string{"llm", "gemini"}, "api_key", apiKey)
		config.SetNestedValue(cfgMap, []string{"llm", "gemini"}, "model", model)
	case providerCustom:
		baseURL, err := promptRequired("Enter base URL:")
		if err != nil {
			return err
		}
		displayBaseURL = baseURL
		config.SetNestedValue(cfgMap, []string{"llm"}, "provider", "openai")
		config.SetNestedValue(cfgMap, []string{"llm", "openai"}, "base_url", baseURL)
		config.SetNestedValue(cfgMap, []string{"llm", "openai"}, "api_key", apiKey)
		config.SetNestedValue(cfgMap, []string{"llm", "openai"}, "model", model)
	default:
		return fmt.Errorf("unsupported provider %q", provider)
	}
	if err := config.WriteFileMap(cfgMap); err != nil {
		return err
	}
	bullets := []string{
		fmt.Sprintf("Provider: %s", provider),
		fmt.Sprintf("Model: %s", model),
	}
	if displayBaseURL != "" {
		bullets = append(bullets, fmt.Sprintf("Base URL: %s", displayBaseURL))
	}
	writeStatus("Saved LLM Settings", bullets)
	return nil
}

func newLoginTTSCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tts",
		Short: "Configure the TTS provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			nativeInfo := native.CurrentInfo()
			options := []string{nativeInfo.Label, "inworld"}
			index, err := promptSelect("Select provider:", options, 0)
			if err != nil {
				if errors.Is(err, errPromptAborted) {
					return nil
				}
				return err
			}
			if index == 0 {
				return handleNativeTTS(nativeInfo)
			}
			return handleInworldTTS()
		},
	}
}

func handleInworldTTS() error {
	apiKey, err := promptRequired("Enter Inworld API key:")
	if err != nil {
		if errors.Is(err, errPromptAborted) {
			return nil
		}
		return err
	}
	models := []string{"inworld-tts-1.5-mini", "inworld-tts-1.5-max"}
	index, err := promptSelect("Select Inworld model:", models, 1)
	if err != nil {
		if errors.Is(err, errPromptAborted) {
			return nil
		}
		return err
	}
	cfgMap, err := config.ReadFileMap()
	if err != nil {
		return err
	}
	config.SetNestedValue(cfgMap, []string{"tts"}, "provider", "inworld")
	config.SetNestedValue(cfgMap, []string{"tts", "inworld"}, "api_key", apiKey)
	config.SetNestedValue(cfgMap, []string{"tts", "inworld"}, "model", models[index])
	if err := config.WriteFileMap(cfgMap); err != nil {
		return err
	}
	writeStatus("Saved Inworld TTS Settings", []string{
		fmt.Sprintf("Model: %s", models[index]),
	})
	return nil
}

func handleNativeTTS(nativeInfo native.Info) error {
	if nativeInfo.Executable != "" {
		if _, err := exec.LookPath(nativeInfo.Executable); err != nil {
			if nativeInfo.InstallCommand == "" {
				writeStatus("Native TTS Missing", []string{
					fmt.Sprintf("Tool: %s", nativeInfo.Label),
					"Install command unavailable.",
				})
				return nil
			}
			ok, err := promptYesNo(fmt.Sprintf("Install %s now?", nativeInfo.Label), true)
			if err != nil {
				if errors.Is(err, errPromptAborted) {
					return nil
				}
				return err
			}
			if ok {
				if err := install.Run(nativeInfo.InstallCommand); err != nil {
					writeStatus("Install Failed", []string{
						fmt.Sprintf("Tool: %s", nativeInfo.Label),
						fmt.Sprintf("Error: %v", err),
					})
					printInstallHelp(nativeInfo)
					return nil
				}
				if _, err := exec.LookPath(nativeInfo.Executable); err != nil {
					writeStatus("Native TTS Still Unavailable", []string{
						fmt.Sprintf("Tool: %s", nativeInfo.Label),
					})
					printInstallHelp(nativeInfo)
					return nil
				}
			} else {
				printInstallHelp(nativeInfo)
			}
		}
	}
	if nativeInfo.ReadyMessage != "" {
		writeStatus("Native TTS Ready", []string{
			nativeInfo.ReadyMessage,
		})
		return setNativeTTSProvider()
	}
	writeStatus("Native TTS Ready", []string{
		fmt.Sprintf("Tool: %s", nativeInfo.Label),
	})
	return setNativeTTSProvider()
}

func setNativeTTSProvider() error {
	cfgMap, err := config.ReadFileMap()
	if err != nil {
		return err
	}
	config.SetNestedValue(cfgMap, []string{"tts"}, "provider", "native")
	return config.WriteFileMap(cfgMap)
}

func printInstallHelp(nativeInfo native.Info) {
	if nativeInfo.InstallLink == "" {
		return
	}
	fmt.Fprintln(os.Stdout, "Install Guide")
	fmt.Fprintln(os.Stdout, nativeInfo.InstallLink)
}

func writeStatus(title string, bullets []string) {
	fmt.Fprintln(os.Stdout, title)
	for _, bullet := range bullets {
		fmt.Fprintln(os.Stdout, bullet)
	}
}

func promptModelSelection(provider string) (string, error) {
	if provider == providerCustom {
		return promptRequired("Enter model:")
	}
	var options []string
	switch provider {
	case "openai":
		options = []string{
			"gpt-4o",
			"gpt-4.1-mini",
			"gpt-4.1",
			"gpt-5-nano",
			"gpt-5-mini",
			"gpt-5",
			"gpt-5.4-nano",
			"gpt-5.4-mini",
			"gpt-5.4",
			"gpt-5.4-pro",
			"Custom (enter manually)",
		}
	case "anthropic":
		options = []string{
			"claude-haiku-4-5",
			"claude-sonnet-4-6",
			"claude-opus-4-6",
			"Custom (enter manually)",
		}
	case "gemini":
		options = []string{
			"gemini-2.5-flash-lite",
			"gemini-2.5-flash",
			"gemini-2.5-pro",
			"gemini-3-flash-preview",
			"gemini-3.1-flash-lite-preview",
			"gemini-3.1-pro-preview",
			"Custom (enter manually)",
		}
	default:
		return promptRequired("Enter model:")
	}
	index, err := promptSelect("Select model:", options, 0)
	if err != nil {
		return "", err
	}
	if index == len(options)-1 {
		return promptRequired("Enter model:")
	}
	return options[index], nil
}
