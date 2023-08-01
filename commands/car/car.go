package car

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	tg "mr-weasel/manager/telegram"
	"strconv"
)

type DraftCar struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Year  int64  `db:"year"`
	Plate string `db:"plate"`
}

type CarCommand struct {
	db        *sqlx.DB
	draftCars map[int64]*DraftCar
}

func New(db *sqlx.DB) *CarCommand {
	return &CarCommand{
		db:        db,
		draftCars: make(map[int64]*DraftCar),
	}
}

func (c *CarCommand) Prefix() string {
	return "car"
}

func (c *CarCommand) Description() string {
	return "manage car costs"
}

func (c *CarCommand) Execute(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	switch pl.Command {
	case "/car new":
		return c.newCar(ctx, pl)
	case "/car get":
		return getCar(ctx, pl)
	}
	return listCars(ctx, pl)
}

func (c *CarCommand) newCar(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	return tg.Result{Text: "Please choose a name for your new car.", State: c.newCarName}, nil
}

func (c *CarCommand) newCarName(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	c.draftCars[pl.User.ID] = &DraftCar{Name: pl.Command}
	return tg.Result{Text: "What is the model year?", State: c.newCarYear}, nil
}

func (c *CarCommand) newCarYear(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	year, err := strconv.Atoi(pl.Command)
	if err != nil {
		return tg.Result{Text: "Please enter a valid number.", State: c.newCarYear}, nil
	}
	c.draftCars[pl.User.ID].Year = int64(year)
	return tg.Result{Text: "What is your plate number? /skip", State: c.newCarPlate}, nil
}

func (c *CarCommand) newCarPlate(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	if pl.Command != "/skip" {
		c.draftCars[pl.User.ID].Plate = pl.Command
	}
	car := c.draftCars[pl.User.ID]
	q := "insert into car (user_id, name, year, plate) values (?,?,?,?);"
	_, err := c.db.ExecContext(ctx, q, pl.User.ID, car.Name, car.Year, car.Plate)
	if err != nil {
		return tg.Result{Text: "There is something wrong with your car, please try again."}, err
	}
	res := tg.Result{Text: fmt.Sprintf("New car %s has been created!", car.Name)}
	res.AddKeyboardButton("Back", "/car")
	return res, nil
}

func getCar(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	return tg.Result{Text: "wip"}, nil
}

func listCars(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	res := tg.Result{Text: "Choose your car from the list below:"}
	// for _, car := range cars {
	// 	res.AddKeyboardButton(car.Name, fmt.Sprintf("%s %d", "/car get", car.ID))
	// }
	// res.AddKeyboardRow()
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
