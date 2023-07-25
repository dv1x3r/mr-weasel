package car

import tg "mr-weasel/manager/telegram"

type CarCommand struct{}

func (CarCommand) Prefix() string {
	return "car"
}

func (CarCommand) Description() string {
	return "manage car costs"
}

func (CarCommand) ExecuteTelegram(user tg.User, payload tg.Payload) (tg.Result, error) {
	// if cmd.Args == "me" {
	// 	return tg.Result{Text: "What is your name?", Action: "name"}, nil
	// }
	// if cmd.Action == "name" {
	// 	return tg.Result{Text: "Pong to " + cmd.Args + "!"}, nil
	// }
	return tg.Result{}, nil
}
