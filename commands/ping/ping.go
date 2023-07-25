package ping

import tg "mr-weasel/manager/telegram"

type PingCommand struct{}

func (PingCommand) Prefix() string {
	return "ping"
}

func (PingCommand) Description() string {
	return "answer with pong"
}

func (PingCommand) ExecuteTelegram(user tg.User, payload tg.Payload) (tg.Result, error) {
	if payload.Text == "me" {
		return tg.Result{Text: "What is your name?", Action: "name"}, nil
	}
	if payload.Action == "name" {
		return tg.Result{Text: "Pong to " + payload.Text + "!"}, nil
	}
	return tg.Result{Text: "pong!"}, nil
}
