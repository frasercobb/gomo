package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/blang/semver"
)

type Module struct {
	Name        string
	FromVersion semver.Version
	ToVersion   semver.Version
}

type Discoverer struct {
	Executor        Executor
	ModuleRegex     string
	ListCommand     string
	ListCommandArgs []string
}

func NewDiscoverer(executor Executor) *Discoverer {
	return &Discoverer{
		Executor:    executor,
		ModuleRegex: "(.+): (.+) -> (.+)",
		ListCommand: "go",
		ListCommandArgs: []string{
			"list", "-u", "-f", "'{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}'", "-m", "all",
		},
	}
}

func (d *Discoverer) ListModules() (string, error) {
	output, err := d.Executor.Run(d.ListCommand, d.ListCommandArgs...)
	if err != nil {
		return "", fmt.Errorf("running '%s %s': %w", d.ListCommand, d.ListCommandArgs, err)
	}
	return output, nil
}

func (d *Discoverer) ParseModules(listOutput string) ([]Module, error) {
	re, err := regexp.Compile(d.ModuleRegex)
	if err != nil {
		return nil, err
	}

	var modules []Module
	split := strings.Split(listOutput, "\n")
	for _, moduleLine := range split {
		matches := re.FindStringSubmatch(moduleLine)
		if len(matches) != 4 {
			return nil, fmt.Errorf("regex was not able to find all matches")
		}

		from, err := semver.Parse(matches[2])
		if err != nil {
			return nil, fmt.Errorf("parsing from version: %w", err)
		}

		to, err := semver.Parse(matches[3])
		if err != nil {
			return nil, fmt.Errorf("parsing to version: %w", err)
		}

		modules = append(modules, Module{
			Name:        matches[1],
			FromVersion: from,
			ToVersion:   to,
		})
	}

	return modules, nil
}
