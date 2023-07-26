package car

import tg "mr-weasel/manager/telegram"

type CarCommand struct{}

func (CarCommand) Prefix() string {
	return "car"
}

func (CarCommand) Description() string {
	return "manage car costs"
}

func (CarCommand) ExecuteTelegram(pl tg.Payload) (tg.Result, error) {
	keyboard := &tg.InlineKeyboardMarkup{
		InlineKeyboard: [][]tg.InlineKeyboardButton{
			{{Text: "abc", CallbackData: "1"}},
		},
	}
	return tg.Result{Text: ".", Keyboard: keyboard}, nil
}
