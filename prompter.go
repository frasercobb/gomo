package main

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

type Prompter struct{}

func NewPrompter() *Prompter {
	return &Prompter{}
}

func (p *Prompter) AskForUpgrades(modules []Module) ([]Module, error) {
	options := createSelectOptions(modules)

	prompt := &survey.MultiSelect{
		Message: "Which modules do you want to upgrade?",
		Options: options,
	}

	var choices []int
	if err := survey.AskOne(prompt, &choices); err != nil {
		return nil, fmt.Errorf("unable to create upgrade prompt: %w", err)
	}

	var modulesToUpgrade []Module
	for _, choice := range choices {
		modulesToUpgrade = append(modulesToUpgrade, modules[choice])
	}

	return modulesToUpgrade, nil
}

func createSelectOptions(modules []Module) []string {
	color.NoColor = false
	var result []string
	for _, mod := range modules {
		result = append(result, moduleToSelectPrompt(mod))
	}
	return result
}

func moduleToSelectPrompt(mod Module) string {
	result := fmt.Sprintf("%s %s -> %s", mod.Name, mod.FromVersion, mod.ToVersion)
	if mod.PatchUpgrade {
		result = color.GreenString(result)
	}

	if mod.MinorUpgrade {
		result = color.BlueString(result)
	}
	return result
}
