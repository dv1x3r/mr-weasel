package ping

import tg "mr-weasel/manager/telegram"

type PingCommand struct{}

func (PingCommand) Prefix() string {
	return "ping"
}

func (PingCommand) Description() string {
	return "answers with pong!"
}

func (PingCommand) ExecuteTelegram(input tg.Input) (tg.Output, error) {
	return tg.Output{Text: "pong!"}, nil
}
