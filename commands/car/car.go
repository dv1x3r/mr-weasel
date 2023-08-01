package car

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	tg "mr-weasel/manager/telegram"
)

type CarCommand struct {
	draftCar map[int64]Car
}

func (CarCommand) Prefix() string {
	return "car"
}

func (CarCommand) Description() string {
	return "manage car costs"
}

func (CarCommand) Execute(ctx context.Context, db *sqlx.DB, pl tg.Payload) (tg.Result, error) {
	switch pl.Command {
	case "/car new":
		return newCar(ctx, db, pl)
	case "/car get":
		return getCar(ctx, db, pl)
	}
	return listCars(ctx, db, pl)
}

type Car struct {
	ID   int64
	Name string
}

var cars = []Car{{ID: 1, Name: "Lexus IS250"}, {ID: 2, Name: "BMW 520i"}}

func newCar(ctx context.Context, db *sqlx.DB, pl tg.Payload) (tg.Result, error) {
	res := tg.Result{
		Text:  "Please choose a name for your new car.",
		State: newCarName,
	}
	return res, nil
}

func newCarName(ctx context.Context, db *sqlx.DB, pl tg.Payload) (tg.Result, error) {
	cars = append(cars, Car{ID: 3, Name: pl.Command})
	res := tg.Result{Text: fmt.Sprintf("New car %s has been created!", pl.Command)}
	res.AddKeyboardButton("Back", "/car")
	return res, nil
}

func getCar(ctx context.Context, db *sqlx.DB, pl tg.Payload) (tg.Result, error) {
	return tg.Result{Text: "wip"}, nil
}

func listCars(ctx context.Context, db *sqlx.DB, pl tg.Payload) (tg.Result, error) {
	res := tg.Result{Text: "Choose your car from the list below:"}
	for _, car := range cars {
		res.AddKeyboardButton(car.Name, fmt.Sprintf("%s %d", "/car get", car.ID))
	}
	res.AddKeyboardRow()
	res.AddKeyboardButton("New", "/car new")
	return res, nil
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
