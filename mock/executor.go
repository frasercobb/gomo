package mock

import "strings"

type Executor struct {
	RunError      error
	RunCalls      []RunCall
	CommandOutput string
}

type RunCall struct {
	Command string
	Args    string
}

func (e *Executor) Run(command string, commandArgs ...string) (string, error) {
	e.RunCalls = append(e.RunCalls, RunCall{
		Command: command,
		Args:    strings.Join(commandArgs, " "),
	})

	return e.CommandOutput, e.RunError
}
