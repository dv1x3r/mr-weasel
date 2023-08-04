package car

import (
	"context"
	"fmt"
	tgclient "mr-weasel/client/telegram"
	tg "mr-weasel/manager/telegram"
	st "mr-weasel/storage"
	"strconv"
	"strings"
)

type CarCommand struct {
	storage   *st.CarStorage
	draftCars map[int64]*st.Car
}

func New(storage *st.CarStorage) *CarCommand {
	return &CarCommand{
		storage:   storage,
		draftCars: make(map[int64]*st.Car),
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
	// if strings.HasPrefix(pl.Command, "/car upd") {
	// 	return c.updCar(ctx, pl)
	// }
	if strings.HasPrefix(pl.Command, "/car del") {
		return c.delCar(ctx, pl)
	}
	return c.selectCars(ctx, pl)
}

func (c *CarCommand) selectCars(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	res := tg.Result{Text: "Choose your car from the list below:", Keyboard: &tgclient.InlineKeyboardMarkup{}}
	cars, err := c.storage.SelectCars(ctx, pl.User.ID)
	if err != nil {
		return tg.Result{Text: "There is something wrong with our database, please try again."}, err
	}
	for i, v := range cars {
		res.Keyboard.AddButton(v.Name, fmt.Sprintf("/car get %d", v.ID))
		if (i+1)%2 == 0 {
			res.Keyboard.AddRow()
		}
	}
	res.Keyboard.AddRow()
	res.Keyboard.AddButton("New", "/car new")
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
			car.ID, car.Name),
		Keyboard: &tgclient.InlineKeyboardMarkup{},
	}

	res.Keyboard.AddButton("Delete", fmt.Sprintf("/car del %d", id))
	res.Keyboard.AddButton("Back", "/car")
	return res, nil
}

func (c *CarCommand) delCar(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	args, _ := strings.CutPrefix(pl.Command, "/car del ")
	id, err := strconv.Atoi(args)
	if err != nil {
		return tg.Result{Text: "There is something wrong with our database, please try again."}, err
	}
	car, err := c.storage.GetCar(ctx, pl.User.ID, int64(id))
	if err != nil {
		return tg.Result{Text: "There is something wrong with our database, please try again."}, err
	}
	c.draftCars[pl.User.ID] = &car
	return tg.Result{Text: fmt.Sprintf("Are you sure you want to delete %s (%d)? /confirm", car.Name, car.Year), State: c.delCarConfirm}, nil
}

func (c *CarCommand) delCarConfirm(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	if pl.Command != "/confirm" {
		return tg.Result{Text: "Operation has been cancelled."}, nil
	}
	affected, err := c.storage.DeleteCar(ctx, pl.User.ID, c.draftCars[pl.User.ID].ID)
	if err != nil {
		return tg.Result{Text: "There is something wrong with our database, please try again."}, err
	}
	if affected == 1 {
		res := tg.Result{Text: "Car has been successfully deleted!", Keyboard: &tgclient.InlineKeyboardMarkup{}}
		res.Keyboard.AddButton("Back", "/car")
		return res, nil
	}
	return tg.Result{Text: "Car not found."}, nil
}

func (c *CarCommand) newCar(ctx context.Context, pl tg.Payload) (tg.Result, error) {
	c.draftCars[pl.User.ID] = &st.Car{UserID: pl.User.ID}
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
		plate := pl.Command
		c.draftCars[pl.User.ID].Plate = &plate
	}
	id, err := c.storage.InsertCar(ctx, *c.draftCars[pl.User.ID])
	if err != nil {
		return tg.Result{Text: "There is something wrong with our database, please try again."}, err
	}
	pl.Command = fmt.Sprintf("/car get %d", id)
	return c.getCar(ctx, pl)
}
