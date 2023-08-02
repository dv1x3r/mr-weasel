package car

import (
	"context"
	"fmt"
	tg "mr-weasel/manager/telegram"
	"mr-weasel/storage"
	"strconv"
	"strings"
)

type DraftCar = storage.Car

type CarCommand struct {
	storage   *storage.CarStorage
	draftCars map[int64]*DraftCar
}

func New(storage *storage.CarStorage) *CarCommand {
	return &CarCommand{
		storage:   storage,
		draftCars: make(map[int64]*DraftCar),
	}
}

func (CarCommand) Prefix() string {
	return "car"
}

func (CarCommand) Description() string {
	return "manage car costs"
}

func (c *CarCommand) Execute(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	if strings.HasPrefix(pl.Command, "/car new") {
		return c.newCar(ctx, pl)
	}
	if strings.HasPrefix(pl.Command, "/car get") {
		return c.getCar(ctx, pl)
	}
	return c.selectCars(ctx, pl)
}

func (c *CarCommand) selectCars(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	res := tg.Result{Text: "Choose your car from the list below:"}
	cars, err := c.storage.SelectCars(ctx, pl.User.ID)
	if err != nil {
		return tg.Result{Text: "There is something wrong with our database, please try again."}, err
	}
	for i, v := range cars {
		res.AddKeyboardButton(v.Name, fmt.Sprintf("%s %d", "/car get", v.ID))
		if (i+1)%2 == 0 {
			res.AddKeyboardRow()
		}
	}
	res.AddKeyboardRow()
	res.AddKeyboardButton("New", "/car new")
	return res, nil
}

func (c *CarCommand) getCar(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	args, _ := strings.CutPrefix(pl.Command, "/car get ")
	id, err := strconv.Atoi(args)
	if err != nil {
		return tg.Result{Text: "There is something wrong with our database, please try again."}, err
	}

	car, err := c.storage.GetCar(ctx, pl.User.ID, int64(id))
	res := tg.Result{
		Text: fmt.Sprintf("*Car ID*: %d \n %s\n",
			car.ID, car.Name)}
	res.AddKeyboardButton("Back", "/car")
	return res, nil
}

func (c *CarCommand) newCar(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	c.draftCars[pl.User.ID] = &DraftCar{UserID: pl.User.ID}
	return tg.Result{Text: "Please choose a name for your new car.", State: c.newCarName}, nil
}

func (c *CarCommand) newCarName(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	c.draftCars[pl.User.ID].Name = pl.Command
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
		*c.draftCars[pl.User.ID].Plate = pl.Command
	}

	id, err := c.storage.InsertCar(ctx, *c.draftCars[pl.User.ID])
	if err != nil {
		return tg.Result{Text: "There is something wrong with our database, please try again."}, err
	}

	pl.Command = fmt.Sprintf("/car get %d", id)
	return c.getCar(ctx, pl)
}
