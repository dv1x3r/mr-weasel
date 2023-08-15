package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	st "mr-weasel/storage"
	"strconv"
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
	return "/car"
}

func (CarCommand) Description() string {
	return "manage car costs"
}

func (c *CarCommand) Execute(ctx context.Context, pl Payload) (Result, error) {
	// if pl.Command == "/car add" {
	// 	return c.addCarStart(ctx, pl)
	// } else if strings.HasPrefix(pl.Command, "/car get ") {
	// 	// return c.getCar(ctx, pl.UserID, pl.)
	// } else if strings.HasPrefix(pl.Command, "/car upd ") {
	// 	// return c.updCar(ctx, pl)
	// } else if strings.HasPrefix(pl.Command, "/car del ") {
	// 	// return c.delCar(ctx, pl)
	// } else if strings.HasPrefix(pl.Command, "/car rmf ") {
	// 	// return c.delCarConfirmed(ctx, pl)
	// }
	return c.showCarList(ctx, pl.UserID)
}

func (c *CarCommand) setDraftCarNew(userID int64) {
	c.draftCars[userID] = &st.Car{UserID: userID}
}

func (c *CarCommand) setDraftCarName(userID int64, input string) {
	c.draftCars[userID].Name = input
}

func (c *CarCommand) setDraftCarYear(userID int64, input string) error {
	year, err := strconv.Atoi(input)
	c.draftCars[userID].Year = int64(year)
	return err
}

func (c *CarCommand) setDraftCarPlate(userID int64, input string) {
	if input != "/skip" {
		c.draftCars[userID].Plate = &input
	}
}

// Add

func (c *CarCommand) addCarStart(ctx context.Context, pl Payload) (Result, error) {
	c.setDraftCarNew(pl.UserID)
	return Result{Text: "Please choose a name for your car.", State: c.addCarName}, nil
}

func (c *CarCommand) addCarName(ctx context.Context, pl Payload) (Result, error) {
	c.setDraftCarName(pl.UserID, pl.Command)
	return Result{Text: "What is the model year?", State: c.addCarYear}, nil
}

func (c *CarCommand) addCarYear(ctx context.Context, pl Payload) (Result, error) {
	err := c.setDraftCarYear(pl.UserID, pl.Command)
	if err != nil {
		return Result{Text: "Please enter a valid number.", State: c.addCarYear}, nil
	}
	return Result{Text: "What is your plate number? /skip", State: c.addCarPlate}, nil
}

func (c *CarCommand) addCarPlate(ctx context.Context, pl Payload) (Result, error) {
	c.setDraftCarPlate(pl.UserID, pl.Command)
	id, err := c.storage.InsertCarIntoDB(ctx, *c.draftCars[pl.UserID])
	if err != nil {
		return Result{Text: "There is something wrong with our database, please try again."}, err
	}
	return c.showCarInfo(ctx, pl.UserID, id)
}

// Select

func (c *CarCommand) showCarList(ctx context.Context, userID int64) (Result, error) {
	cars, err := c.storage.SelectCarsFromDB(ctx, userID)
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
	res.AddKeyboardButton("« Add new Car »", "/car add")
	return res, nil
}

func (c *CarCommand) showCarInfo(ctx context.Context, userID int64, carID int64) (Result, error) {
	car, err := c.storage.GetCarFromDB(ctx, userID, carID)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{Text: "Car not found."}, nil
	} else if err != nil {
		return Result{Text: "There is something wrong with our database, please try again."}, err
	}

	res := Result{}
	res.Text += fmt.Sprintf("<b>Name:</b> %s\n", car.Name)
	res.Text += fmt.Sprintf("<b>Year:</b> %d\n", car.Year)
	if car.Plate != nil {
		res.Text += fmt.Sprintf("<b>Plates:</b> %s\n", *car.Plate)
	} else {
		res.Text += fmt.Sprintf("<b>Plates:</b> 🚫\n")
	}
	res.AddKeyboardButton("Fuel", fmt.Sprintf("/car fuel %d", carID))
	res.AddKeyboardButton("Edit Car", fmt.Sprintf("/car upd %d", carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Service", fmt.Sprintf("/car service %d", carID))
	res.AddKeyboardButton("Delete Car", fmt.Sprintf("/car del %d", carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("« Back to Cars list", "/car")
	return res, nil
}

// Delete

func (c *CarCommand) delCar(ctx context.Context, userID int64, carID int64) (Result, error) {
	car, err := c.storage.GetCarFromDB(ctx, userID, carID)
	if err != nil {
		return Result{Text: "Car not found."}, err
	}
	res := Result{Text: fmt.Sprintf("Are you sure you want to delete %s (%d)?", car.Name, car.Year)}
	res.AddKeyboardButton("Yes, delete the car", fmt.Sprintf("/car rm %d", carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("No", fmt.Sprintf("/car get %d", carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Nope, nevermind", fmt.Sprintf("/car get %d", carID))
	return res, nil
}

func (c *CarCommand) delCarConfirmed(ctx context.Context, userID int64, carID int64) (Result, error) {
	affected, err := c.storage.DeleteCarFromDB(ctx, userID, carID)
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
