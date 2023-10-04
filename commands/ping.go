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

func (PingCommand) Execute(ctx context.Context, pl Payload) {
	if pl.Command == "/ping me" {
		pl.ResultChan <- Result{Text: "What is your name?", State: personalized}
	} else {
		pl.ResultChan <- Result{Text: "pong!"}
	}
}

func personalized(ctx context.Context, pl Payload) {
	pl.ResultChan <- Result{Text: "Pong to " + _es(pl.Command) + "!"}
}
