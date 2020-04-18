package main

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/fatih/color"
)

type Prompter struct {
	stdio terminal.Stdio
}

type PrompterOption func(*Prompter)

func NewPrompter(options ...PrompterOption) *Prompter {
	p := &Prompter{}

	for _, option := range options {
		option(p)
	}

	return p
}

func WithStdio(stdio terminal.Stdio) PrompterOption {
	return func(p *Prompter) {
		p.stdio = stdio
	}
}

func (p *Prompter) AskForUpgrades(modules []Module) ([]Module, error) {
	options := createSelectOptions(modules)

	prompt := &survey.MultiSelect{
		Message: "Which modules do you want to upgrade?",
		Options: options,
	}

	var choices []int
	if err := survey.AskOne(prompt, &choices, survey.WithStdio(p.stdio.In, p.stdio.Out, p.stdio.Err)); err != nil {
		return nil, fmt.Errorf("unable to get module choices: %w", err)
	}

	var modulesToUpgrade []Module
	for _, choice := range choices {
		modulesToUpgrade = append(modulesToUpgrade, modules[choice])
	}

	return modulesToUpgrade, nil
}

func createSelectOptions(modules []Module) []string {
	color.NoColor = false // https://github.com/golang/go/issues/18153
	var minorUpgrades []Module
	var patchUpgrades []Module
	for _, m := range modules {
		if m.MinorUpgrade {
			minorUpgrades = append(minorUpgrades, m)
		}
		if m.PatchUpgrade {
			patchUpgrades = append(patchUpgrades, m)
		}
	}

	var result []string
	for _, mod := range patchUpgrades {
		result = append(result, moduleToSelectPrompt(mod))
	}
	for _, mod := range minorUpgrades {
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
