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

/*

/car

Traveled in the past year: 10,000 KM (+2%)
Gas Expenses: 1000 EUR (+10%)
Other Expenses: 500 EUR (-10%)
License expiration: 01-Jan-2099

Please select your car for additional actions:

| Lexus IS250 (2011) | BMW 520i (2021) |
| Previous | Next |


/car 1

Car name: BMW 520i (2021)
Traveled in the past year: 10,000 KM (+2%)
Gas Expenses: 1000 EUR (+10%)
Other Expenses: 500 EUR (-10%)

| Add Gas | Add Service | Update Attributes
| Edit Gas | Edit Service |

*/
