package car

import (
	tg "mr-weasel/manager/telegram"
)

type CarCommand struct{}

func (CarCommand) Prefix() string {
	return "car"
}

func (CarCommand) Description() string {
	return "manage car costs"
}

func (CarCommand) ExecuteTelegram(pl tg.Payload) (tg.Result, error) {
	switch pl.Command.Action {
	case "new":
		newCar()
	default:
		listCars()
	}
	// keyboard := &tg.InlineKeyboardMarkup{
	// 	InlineKeyboard: [][]tg.InlineKeyboardButton{
	// 		{{Text: "abc", CallbackData: "1"}},
	// 	},
	// }
	// return tg.Result{Text: ".", Keyboard: keyboard}, nil
	return tg.Result{}, nil
}

func listCars() {

}

func newCar() {

}

/*

/car

Traveled in the past year: 10,000 KM (+2%)
Fuel consumption: 10.0L/100Km
Fuel expenses: 1000 EUR (+10%)
Other expenses: 500 EUR (-10%)

Choose your car from the list below:

| Lexus IS250 (2011) | BMW 520i (2021) |
| Previous | Next | NEW |


/car:newcar

Please choose a name for your new car.

/car 1

BMW 520i (2021)

| Add Gas | Add Service |
| Edit Gas | Edit Service |

*/
