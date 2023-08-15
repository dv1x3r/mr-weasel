package commands

import "context"

type PingCommand struct{}

func NewPingCommand() *PingCommand {
	return &PingCommand{}
}

func (PingCommand) Prefix() string {
	return "/ping"
}

func (PingCommand) Description() string {
	return "answer with pong"
}

func (PingCommand) Execute(ctx context.Context, pl Payload) (Result, error) {
	if pl.Command == "/ping me" {
		return Result{Text: "What is your name?", State: personalized}, nil
	}
	return Result{Text: "pong!"}, nil
}

func personalized(ctx context.Context, pl Payload) (Result, error) {
	return Result{Text: "Pong to " + pl.Command + "!"}, nil
}
