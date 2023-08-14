package commands

import (
	"context"
	"fmt"
	"mr-weasel/storage"
	st "mr-weasel/storage"
	"strconv"
	"strings"
)

type CarCommand struct {
	storage   *st.CarStorage
	draftCars map[int64]*st.Car
}

func NewCarCommand(storage *st.CarStorage) *CarCommand {
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

func (c *CarCommand) Execute(ctx context.Context, pl Payload) (Result, error) {
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
	return c.selectCarsFromDB(ctx, pl)
}

func (c *CarCommand) selectCarsFromDB(ctx context.Context, pl Payload) (Result, error) {
	cars, err := c.storage.SelectCars(ctx, pl.UserID)
	if err != nil {
		return Result{Text: "There is something wrong with our database, please try again."}, err
	}
	res := c.addCarsToResult(cars)
	return res, nil
}

func (c *CarCommand) addCarsToResult(cars []storage.Car) Result {
	res := Result{Text: "Choose your car from the list below:"}
	for i, v := range cars {
		res.AddKeyboardButton(v.Name, fmt.Sprintf("/car get %d", v.ID))
		if (i+1)%2 == 0 {
			res.AddKeyboardRow()
		}
	}
	res.AddKeyboardRow()
	res.AddKeyboardButton("New", "/car new")
	return res
}

func (c *CarCommand) getCar(ctx context.Context, pl Payload) (Result, error) {
	args, _ := strings.CutPrefix(pl.Command, "/car get ")
	id, err := strconv.Atoi(args)
	if err != nil {
		return Result{Text: "Invalid Car ID."}, err
	}

	car, err := c.storage.GetCar(ctx, pl.UserID, int64(id))
	if err != nil {
		return Result{Text: "Car not found."}, err
	}

	res := Result{Text: fmt.Sprintf("*Car ID*: %d \n %s\n", car.ID, car.Name)}
	res.AddKeyboardButton("Delete", fmt.Sprintf("/car del %d", id))
	res.AddKeyboardButton("Back", "/car")
	return res, nil
}

func (c *CarCommand) delCar(ctx context.Context, pl Payload) (Result, error) {
	args, _ := strings.CutPrefix(pl.Command, "/car del ")
	id, err := strconv.Atoi(args)
	if err != nil {
		return Result{Text: "There is something wrong with our database, please try again."}, err
	}
	car, err := c.storage.GetCar(ctx, pl.UserID, int64(id))
	if err != nil {
		return Result{Text: "There is something wrong with our database, please try again."}, err
	}
	c.draftCars[pl.UserID] = &car
	return Result{Text: fmt.Sprintf("Are you sure you want to delete %s (%d)? /confirm", car.Name, car.Year), State: c.delCarConfirm}, nil
}

func (c *CarCommand) delCarConfirm(ctx context.Context, pl Payload) (Result, error) {
	if pl.Command != "/confirm" {
		return Result{Text: "Operation has been cancelled."}, nil
	}
	affected, err := c.storage.DeleteCar(ctx, pl.UserID, c.draftCars[pl.UserID].ID)
	if err != nil {
		return Result{Text: "There is something wrong with our database, please try again."}, err
	}
	if affected == 1 {
		res := Result{Text: "Car has been successfully deleted!"}
		res.AddKeyboardButton("Back", "/car")
		return res, nil
	}
	return Result{Text: "Car not found."}, nil
}

func (c *CarCommand) newCar(ctx context.Context, pl Payload) (Result, error) {
	c.draftCars[pl.UserID] = &st.Car{UserID: pl.UserID}
	return Result{Text: "Please choose a name for your new car.", State: c.newCarName}, nil
}

func (c *CarCommand) newCarName(ctx context.Context, pl Payload) (Result, error) {
	c.draftCars[pl.UserID].Name = pl.Command
	return Result{Text: "What is the model year?", State: c.newCarYear}, nil
}

func (c *CarCommand) newCarYear(ctx context.Context, pl Payload) (Result, error) {
	year, err := strconv.Atoi(pl.Command)
	if err != nil {
		return Result{Text: "Please enter a valid number.", State: c.newCarYear}, nil
	}
	c.draftCars[pl.UserID].Year = int64(year)
	return Result{Text: "What is your plate number? /skip", State: c.newCarPlate}, nil
}

func (c *CarCommand) newCarPlate(ctx context.Context, pl Payload) (Result, error) {
	if pl.Command != "/skip" {
		plate := pl.Command
		c.draftCars[pl.UserID].Plate = &plate
	}
	id, err := c.storage.InsertCar(ctx, *c.draftCars[pl.UserID])
	if err != nil {
		return Result{Text: "There is something wrong with our database, please try again."}, err
	}
	pl.Command = fmt.Sprintf("/car get %d", id)
	return c.getCar(ctx, pl)
}
