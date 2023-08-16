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
	args := splitCommand(pl.Command, c.Prefix())
	subcommand, carID := safeGet(args, 0), safeGetInt(args, 1)
	switch subcommand {
	case "add":
		return c.addCarStart(ctx, pl)
	case "get":
		return c.carMainMenu(ctx, pl.UserID, carID)
	case "upd":
		if field := safeGet(args, 2); field != "" {
			return c.updCarField(ctx, pl.UserID, carID, field)
		} else {
			return c.carUpdateMenu(ctx, pl.UserID, carID)
		}
	case "del":
		return c.delCarCheck(ctx, pl.UserID, carID)
	case "rmf":
		return c.delCarConfirmed(ctx, pl.UserID, carID)
	default:
		return c.carsList(ctx, pl.UserID)
	}
}

func (c *CarCommand) draftCarFetchFromDB(ctx context.Context, userID int64, carID int64) error {
	car, err := c.storage.GetCarFromDB(ctx, userID, carID)
	if err != nil {
		return err
	}
	c.draftCars[userID] = &car
	return nil
}

func (c *CarCommand) draftCarInsertIntoDB(ctx context.Context, userID int64) (int64, error) {
	return c.storage.InsertCarIntoDB(ctx, *c.draftCars[userID])
}

func (c *CarCommand) draftCarUpdateInDB(ctx context.Context, userID int64) (int64, error) {
	return c.storage.UpdateCarInDB(ctx, *c.draftCars[userID])
}

func (c *CarCommand) draftCarInit(userID int64) {
	c.draftCars[userID] = &st.Car{UserID: userID}
}

func (c *CarCommand) draftCarSetName(userID int64, input string) {
	c.draftCars[userID].Name = input
}

func (c *CarCommand) draftCarSetYear(userID int64, input string) error {
	year, err := strconv.Atoi(input)
	c.draftCars[userID].Year = int64(year)
	return err
}

func (c *CarCommand) draftCarSetPlate(userID int64, input string) {
	if input != "/skip" {
		c.draftCars[userID].Plate = &input
	}
}

func (c *CarCommand) addCarStart(ctx context.Context, pl Payload) (Result, error) {
	c.draftCarInit(pl.UserID)
	return Result{Text: "Please choose a name for your car.", State: c.addCarName}, nil
}

func (c *CarCommand) addCarName(ctx context.Context, pl Payload) (Result, error) {
	c.draftCarSetName(pl.UserID, pl.Command)
	return Result{Text: "What is the model year?", State: c.addCarYear}, nil
}

func (c *CarCommand) addCarYear(ctx context.Context, pl Payload) (Result, error) {
	err := c.draftCarSetYear(pl.UserID, pl.Command)
	if err != nil {
		return Result{Text: "Please enter a valid number.", State: c.addCarYear}, nil
	}
	return Result{Text: "What is your plate number? /skip", State: c.addCarPlateAndSave}, nil
}

func (c *CarCommand) addCarPlateAndSave(ctx context.Context, pl Payload) (Result, error) {
	c.draftCarSetPlate(pl.UserID, pl.Command)
	carID, err := c.draftCarInsertIntoDB(ctx, pl.UserID)
	if err != nil {
		return Result{Text: "There is something wrong with our database, please try again."}, err
	}
	return c.carMainMenu(ctx, pl.UserID, carID)
}

func (c *CarCommand) carsList(ctx context.Context, userID int64) (Result, error) {
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

func (c *CarCommand) carInfo(ctx context.Context, userID int64, carID int64) (Result, error) {
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
		res.Text += fmt.Sprintf("<b>Plate:</b> %s\n", *car.Plate)
	} else {
		res.Text += fmt.Sprintf("<b>Plate:</b> 🚫\n")
	}

	return res, nil
}

func (c *CarCommand) carMainMenu(ctx context.Context, userID int64, carID int64) (Result, error) {
	res, err := c.carInfo(ctx, userID, carID)
	if err != nil {
		return res, err
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

func (c *CarCommand) carUpdateMenu(ctx context.Context, userID int64, carID int64) (Result, error) {
	res, err := c.carInfo(ctx, userID, carID)
	if err != nil {
		return res, err
	}
	res.AddKeyboardButton("Change Name", fmt.Sprintf("/car upd %d name", carID))
	res.AddKeyboardButton("Change Year", fmt.Sprintf("/car upd %d year", carID))
	res.AddKeyboardButton("Change Plate", fmt.Sprintf("/car upd %d plate", carID))
	res.AddKeyboardButton("« Back to Car menu", fmt.Sprintf("/car %d", carID))
	return res, nil
}

func (c *CarCommand) updCarField(ctx context.Context, userID int64, carID int64, field string) (Result, error) {
	err := c.draftCarFetchFromDB(ctx, userID, carID)
	if err != nil {
		return Result{Text: "Car not found."}, err
	}
	switch field {
	case "name":
		return Result{Text: "What is the new car name?", State: c.updCarNameAndSave}, nil
	// case "year":
	// 	return Result{Text: "What is the new car year?", State: c.updCarNameAndSave}, nil
	// case "plate":
	// 	return Result{Text: "What is the new car plate? /skip", State: c.updCarNameAndSave}, nil
	default:
		return Result{Text: "Unknown field."}, nil
	}
}

func (c *CarCommand) updCarNameAndSave(ctx context.Context, pl Payload) (Result, error) {
	c.draftCarSetName(pl.UserID, pl.Command)
	_, err := c.draftCarUpdateInDB(ctx, pl.UserID)
	if err != nil {
		return Result{Text: "Update failed, try again."}, err
	}
	res := Result{Text: "Car Name has been successfully updated!"}
	res.AddKeyboardButton("« Back to Car menu", fmt.Sprintf("/car %d", c.draftCars[pl.UserID].ID))
	return res, nil
}

func (c *CarCommand) delCarCheck(ctx context.Context, userID int64, carID int64) (Result, error) {
	car, err := c.storage.GetCarFromDB(ctx, userID, carID)
	if err != nil {
		return Result{Text: "Car not found."}, err
	}

	res := Result{Text: fmt.Sprintf("Are you sure you want to delete %s (%d)?", car.Name, car.Year)}
	res.AddKeyboardButton("Yes, delete the car", fmt.Sprintf("/car rmf %d", carID))
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
