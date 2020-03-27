package main

import "strings"

type MockExecutor struct {
	RunError      error
	RunCalls      []RunCall
	CommandOutput string
}

type RunCall struct {
	Command string
	Args    string
}

func (e *MockExecutor) Run(command string, commandArgs ...string) (string, error) {
	e.RunCalls = append(e.RunCalls, RunCall{
		Command: command,
		Args:    strings.Join(commandArgs, " "),
	})

	return e.CommandOutput, e.RunError
}
