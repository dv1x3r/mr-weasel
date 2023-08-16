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
	c := &CarCommand{
		storage:   storage,
		draftCars: make(map[int64]*st.Car),
	}
	return c
}

func (CarCommand) Prefix() string {
	return "/car"
}

func (CarCommand) Description() string {
	return "manage car costs"
}

func (c *CarCommand) Execute(ctx context.Context, pl Payload) (Result, error) {
	args := splitCommand(pl.Command, c.Prefix())
	switch safeGet(args, 0) {
	case cmdCarAdd:
		return c.addCarStart(ctx, pl)
	case cmdCarGet:
		return c.showCarMenu(ctx, pl.UserID, safeGetInt(args, 1))
	case cmdCarUpd:
		return c.carUpdateMenu(ctx, pl.UserID, safeGetInt(args, 1))
	case cmdCarUpdName:
		return c.carUpdateAskName(ctx, pl.UserID, safeGetInt(args, 1))
	case cmdCarUpdYear:
		return c.carUpdateAskYear(ctx, pl.UserID, safeGetInt(args, 1))
	case cmdCarUpdPlate:
		return c.carUpdateAskPlate(ctx, pl.UserID, safeGetInt(args, 1))
	case cmdCarDel:
		return c.delCarAsk(ctx, pl.UserID, safeGetInt(args, 1))
	case cmdCarDelYes:
		return c.delCarYes(ctx, pl.UserID, safeGetInt(args, 1))
	default:
		return c.showCarList(ctx, pl.UserID)
	}
}

const (
	cmdCarAdd      = "add"
	cmdCarGet      = "get"
	cmdCarUpd      = "upd"
	cmdCarUpdName  = "upd_name"
	cmdCarUpdYear  = "upd_year"
	cmdCarUpdPlate = "upd_plate"
	cmdCarDel      = "del"
	cmdCarDelYes   = "del_yes"
	cmdCarFuel     = "fuel"
	cmdCarService  = "service"
)

func (c *CarCommand) formatCarDetails(car st.Car) string {
	html := fmt.Sprintf("<b>Name:</b> %s\n", car.Name)
	html += fmt.Sprintf("<b>Year:</b> %d\n", car.Year)
	if car.Plate != nil {
		html += fmt.Sprintf("<b>Plate:</b> %s\n", *car.Plate)
	} else {
		html += fmt.Sprintf("<b>Plate:</b> 🚫\n")
	}
	return html
}

func (c *CarCommand) showCarMenu(ctx context.Context, userID int64, carID int64) (Result, error) {
	car, err := c.storage.GetCarFromDB(ctx, userID, carID)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{Text: "Car not found."}, nil
	} else if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}

	res := Result{Text: c.formatCarDetails(car)}
	res.AddKeyboardButton("Fuel", commandf(c, cmdCarFuel, carID))
	res.AddKeyboardButton("Edit Car", commandf(c, cmdCarUpd, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Service", commandf(c, cmdCarService, carID))
	res.AddKeyboardButton("Delete Car", commandf(c, cmdCarDel, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("« Back to Cars list", c.Prefix())
	return res, nil
}

func (c *CarCommand) showCarList(ctx context.Context, userID int64) (Result, error) {
	cars, err := c.storage.SelectCarsFromDB(ctx, userID)
	if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}

	res := Result{Text: "Choose your car from the list below:"}
	for i, v := range cars {
		res.AddKeyboardButton(v.Name, commandf(c, cmdCarGet, v.ID))
		if (i+1)%2 == 0 {
			res.AddKeyboardRow()
		}
	}
	res.AddKeyboardRow()
	res.AddKeyboardButton("« Add new Car »", commandf(c, cmdCarAdd, nil))
	return res, nil
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
	return c.showCarMenu(ctx, pl.UserID, carID)
}

func (c *CarCommand) carUpdateMenu(ctx context.Context, userID int64, carID int64) (Result, error) {
	car, err := c.storage.GetCarFromDB(ctx, userID, carID)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{Text: "Car not found."}, nil
	} else if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}

	res := Result{Text: c.formatCarDetails(car)}
	res.AddKeyboardButton("Edit Name", fmt.Sprintf("/car upd %d name", carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Edit Year", fmt.Sprintf("/car upd %d year", carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Edit Plate", fmt.Sprintf("/car upd %d plate", carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("« Back to Car menu", fmt.Sprintf("/car %d", carID))
	return res, nil
}

func (c *CarCommand) carUpdateAskName(ctx context.Context, userID int64, carID int64) (Result, error) {
	if err := c.draftCarFetchFromDB(ctx, userID, carID); err != nil {
		return Result{Text: "Car not found."}, err
	}
	return Result{Text: "What is the new car name?", State: c.updCarNameAndSave}, nil
}

func (c *CarCommand) carUpdateAskYear(ctx context.Context, userID int64, carID int64) (Result, error) {
	if err := c.draftCarFetchFromDB(ctx, userID, carID); err != nil {
		return Result{Text: "Car not found."}, err
	}
	return Result{Text: "What is the new car year?", State: c.updCarYearAndSave}, nil
}

func (c *CarCommand) carUpdateAskPlate(ctx context.Context, userID int64, carID int64) (Result, error) {
	if err := c.draftCarFetchFromDB(ctx, userID, carID); err != nil {
		return Result{Text: "Car not found."}, err
	}
	return Result{Text: "What is the new car plate? /skip", State: c.updCarPlateAndSave}, nil
}

func (c *CarCommand) updCarNameAndSave(ctx context.Context, pl Payload) (Result, error) {
	c.draftCarSetName(pl.UserID, pl.Command)
	if _, err := c.draftCarUpdateInDB(ctx, pl.UserID); err != nil {
		return Result{Text: "Update failed, try again."}, err
	}
	res := Result{Text: "Car name has been successfully updated!"}
	res.AddKeyboardButton("« Back to Car menu", fmt.Sprintf("/car %d", c.draftCars[pl.UserID].ID))
	return res, nil
}

func (c *CarCommand) updCarYearAndSave(ctx context.Context, pl Payload) (Result, error) {
	c.draftCarSetYear(pl.UserID, pl.Command)
	if _, err := c.draftCarUpdateInDB(ctx, pl.UserID); err != nil {
		return Result{Text: "Update failed, try again."}, err
	}
	res := Result{Text: "Car year has been successfully updated!"}
	res.AddKeyboardButton("« Back to Car menu", fmt.Sprintf("/car %d", c.draftCars[pl.UserID].ID))
	return res, nil
}

func (c *CarCommand) updCarPlateAndSave(ctx context.Context, pl Payload) (Result, error) {
	c.draftCarSetPlate(pl.UserID, pl.Command)
	if _, err := c.draftCarUpdateInDB(ctx, pl.UserID); err != nil {
		return Result{Text: "Update failed, try again."}, err
	}
	res := Result{Text: "Car plate has been successfully updated!"}
	res.AddKeyboardButton("« Back to Car menu", fmt.Sprintf("/car %d", c.draftCars[pl.UserID].ID))
	return res, nil
}

func (c *CarCommand) delCarAsk(ctx context.Context, userID int64, carID int64) (Result, error) {
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

func (c *CarCommand) delCarYes(ctx context.Context, userID int64, carID int64) (Result, error) {
	affected, err := c.storage.DeleteCarFromDB(ctx, userID, carID)
	if err != nil || affected != 1 {
		return Result{Text: "Car not found."}, err
	}

	res := Result{Text: "Car has been successfully deleted!"}
	res.AddKeyboardButton("« Back to Cars list", "/car")
	return res, nil

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
	if input == "/skip" {
		c.draftCars[userID].Plate = nil
	} else {
		c.draftCars[userID].Plate = &input
	}
}
