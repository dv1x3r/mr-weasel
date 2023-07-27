package car

import (
	"fmt"
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
	res := tg.Result{}
	switch pl.Command.Action {
	case "new":
		newCar()
	default:
		res.Text = "Choose your car from the list below:"
		for _, car := range listCars() {
			text := car.Name
			data := fmt.Sprintf("%s %d", "/car:select", car.ID)
			res.AddKeyboardButton(0, text, data)
		}
		res.AddKeyboardButton(1, "New Car", "/car:new")
	}
	return res, nil
}

type Car struct {
	ID   int64
	Name string
}

var cars = []Car{{ID: 1, Name: "Lexus IS250"}, {ID: 2, Name: "BMW 520i"}}

func listCars() []Car {
	return cars
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
