package car

import (
	"context"
	"fmt"
	tg "mr-weasel/manager/telegram"
	"strconv"
	// "strings"
	"github.com/jmoiron/sqlx"
)

type DraftCar struct {
	Name  string
	Year  int64
	Plate string
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
	// split := strings.SplitN(pl.Command, 3)
	switch pl.Command {
	case "/car new":
		return c.newCar(ctx, pl)
	case "/car get":
		return c.getCar(ctx, pl)
	}
	return c.listCars(ctx, pl)
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
	res, err := c.db.ExecContext(ctx, q, pl.User.ID, car.Name, car.Year, car.Plate)
	if err != nil {
		return tg.Result{Text: "There is something wrong with our database, please try again."}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return tg.Result{Text: "There is something wrong with our database, please try again."}, err
	}
	pl.Command = fmt.Sprintf("/car get %d", id)
	return c.getCar(ctx, pl)
}

func (c *CarCommand) listCars(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	res := tg.Result{Text: "Choose your car from the list below:"}
	rows, err := c.db.Query(`
		select id, name || ' (' || year || ')' as desc
		from car
		where user_id = ?
		order by year desc, name`,
		pl.User.ID,
	)
	if err != nil {
		return tg.Result{Text: "There is something wrong with our database, please try again."}, err
	}
	i := 0
	for rows.Next() {
		var id int64
		var name string
		rows.Scan(&id, &name)
		res.AddKeyboardButton(name, fmt.Sprintf("%s %d", "/car get", id))
		if i++; i%2 == 0 {
			res.AddKeyboardRow()
		}
	}
	if err := rows.Err(); err != nil {
		return tg.Result{Text: "There is something wrong with our database, please try again."}, err
	}
	res.AddKeyboardRow()
	res.AddKeyboardButton("New", "/car new")
	return res, nil
}

func (c *CarCommand) getCar(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	res := tg.Result{Text: fmt.Sprintf(`
			Car ID: 
			Car Name: %s
		`, pl.Command)}
	res.AddKeyboardButton("Back", "/car")
	return res, nil
}
