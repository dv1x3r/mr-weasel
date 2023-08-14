package commands

import (
	"context"
	"fmt"
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
	if pl.Command == "/car new" {
		return c.newCarStart(ctx, pl)
	}
	if strings.HasPrefix(pl.Command, "/car get ") {
		return c.getCar(ctx, pl)
	}
	if strings.HasPrefix(pl.Command, "/car upd ") {
		// return c.updCar(ctx, pl)
	}
	if strings.HasPrefix(pl.Command, "/car del ") {
		return c.delCarFromDB(ctx, pl)
	}
	if strings.HasPrefix(pl.Command, "/car rm ") {
		return c.rmCarFromDB(ctx, pl)
	}
	return c.selectCars(ctx, pl)
}

func (c *CarCommand) newCarStart(ctx context.Context, pl Payload) (Result, error) {
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

func (c *CarCommand) selectCars(ctx context.Context, pl Payload) (Result, error) {
	cars, err := c.storage.SelectCars(ctx, pl.UserID)
	if err != nil {
		return Result{Text: "There is something wrong with our database, please try again."}, err
	}
	res := Result{Text: "Choose your car from the list below:"}
	for i, v := range cars {
		res.AddKeyboardButton(v.Name, fmt.Sprintf("/car get %d", v.ID))
		if (i+1)%2 == 0 {
			res.AddKeyboardRow()
		}
	}
	res.AddKeyboardRow()
	res.AddKeyboardButton("« Add new Car »", "/car new")
	return res, nil
}

func (c *CarCommand) getCar(ctx context.Context, pl Payload) (Result, error) {
	args, _ := strings.CutPrefix(pl.Command, "/car get ")
	id, _ := strconv.Atoi(args)
	car, err := c.storage.GetCar(ctx, pl.UserID, int64(id))
	if err != nil {
		return Result{Text: "Car not found."}, err
	}

	res := Result{}
	res.Text += fmt.Sprintf("<b>Name:</b> %s\n", car.Name)
	res.Text += fmt.Sprintf("<b>Year:</b> %d\n", car.Year)
	if car.Plate != nil {
		res.Text += fmt.Sprintf("<b>Plates:</b> %s\n", *car.Plate)
	} else {
		res.Text += fmt.Sprintf("<b>Plates:</b> 🚫\n")
	}

	res.AddKeyboardButton("Fuel", fmt.Sprintf("/car fuel %d", id))
	res.AddKeyboardButton("Edit Car", fmt.Sprintf("/car edit %d", id))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Service", fmt.Sprintf("/car service %d", id))
	res.AddKeyboardButton("Delete Car", fmt.Sprintf("/car del %d", id))
	res.AddKeyboardRow()
	res.AddKeyboardButton("« Back to Cars list", "/car")
	return res, nil
}

func (c *CarCommand) delCarFromDB(ctx context.Context, pl Payload) (Result, error) {
	args, _ := strings.CutPrefix(pl.Command, "/car del ")
	id, _ := strconv.Atoi(args)
	car, err := c.storage.GetCar(ctx, pl.UserID, int64(id))
	if err != nil {
		return Result{Text: "Car not found."}, err
	}

	res := Result{Text: fmt.Sprintf("Are you sure you want to delete %s (%d)?", car.Name, car.Year)}
	res.AddKeyboardButton("Yes, delete the car", fmt.Sprintf("/car rm %d", id))
	res.AddKeyboardRow()
	res.AddKeyboardButton("No", fmt.Sprintf("/car get %d", id))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Nope, nevermind", fmt.Sprintf("/car get %d", id))
	return res, nil
}

func (c *CarCommand) rmCarFromDB(ctx context.Context, pl Payload) (Result, error) {
	args, _ := strings.CutPrefix(pl.Command, "/car rm ")
	id, _ := strconv.Atoi(args)
	affected, err := c.storage.DeleteCar(ctx, pl.UserID, int64(id))
	if err != nil {
		return Result{Text: "Car not found, or you do not have access."}, err
	}
	if affected == 1 {
		res := Result{Text: "Car has been successfully deleted!"}
		res.AddKeyboardButton("« Back to Cars list", "/car")
		return res, nil
	}
	return Result{Text: "Car not found."}, nil
}
