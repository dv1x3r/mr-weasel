package ping

import tg "mr-weasel/manager/telegram"

type PingCommand struct{}

func (PingCommand) Prefix() string {
	return "ping"
}

func (PingCommand) Description() string {
	return "answer with pong"
}

func (PingCommand) Execute(pl tg.Payload) (tg.Result, error) {
	// if pl.Command == "me" {
	// 	return tg.Result{Text: "What is your name?", Action: "name"}, nil
	// }
	// if pl.Command.Action == "name" {
	// 	return tg.Result{Text: "Pong to " + pl.Command.Text + "!"}, nil
	// }
	return tg.Result{Text: "pong!"}, nil
}
