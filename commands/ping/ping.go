package ping

import (
	tg "mr-weasel/manager/telegram"
)

type PingCommand struct{}

func (PingCommand) Prefix() string {
	return "ping"
}

func (PingCommand) Description() string {
	return "answers with pong!"
}

func (PingCommand) ExecuteTelegram(cmd tg.Command) (tg.Result, error) {
	if cmd.Args == "" {
		return tg.Result{Text: "What is your name?", Action: "welcome"}, nil
	}
	if cmd.Action == "welcome" {
		return tg.Result{Text: "Welcome, " + cmd.Args}, nil
	}
	return tg.Result{Text: "pong!"}, nil
}
