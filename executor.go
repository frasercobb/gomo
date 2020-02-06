package main

import (
	"fmt"
	"os/exec"
)

type Executor interface {
	Run(command string, commandArgs ...string) (string, error)
}

type CommandExecutor struct{}

func (c *CommandExecutor) Run(command string, commandArgs ...string) (string, error) {
	cmd := exec.Command(command, commandArgs...)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("executing command %q: %w", fmt.Sprintf("%s %s", command, commandArgs), err)
	}

	return string(output), nil
}
