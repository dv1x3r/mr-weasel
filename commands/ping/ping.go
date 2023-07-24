package ping

import (
	"fmt"
	tg "mr-weasel/manager/telegram"
)

type PingCommand struct{}

func (PingCommand) Prefix() string {
	return "ping"
}

func (PingCommand) Description() string {
	return "answers with pong!"
}

func (PingCommand) ExecuteTelegram(input tg.Input) (tg.Output, error) {
	s := fmt.Sprintf("Pong! %s - %s - %s - %s", input.Prefix, input.Action, input.Args, input.User.Username)
	return tg.Output{Text: s}, nil
}
