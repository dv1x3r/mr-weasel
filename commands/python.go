package commands

import "context"

type PythonCommand struct {
}

func NewPythonCommand() *PythonCommand {
	return &PythonCommand{}
}

func (PythonCommand) Prefix() string {
	return "/python"
}

func (PythonCommand) Description() string {
	return "test long running command"
}

func (c *PythonCommand) Execute(ctx context.Context, pl Payload) (Result, error) {
	return Result{}, nil
}
