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
		return c.carAddStart(ctx, pl)
	case cmdCarGet:
		return c.carShowDetailsMenu(ctx, pl.UserID, safeGetInt(args, 1))
	case cmdCarUpd:
		return c.carShowUpdateMenu(ctx, pl.UserID, safeGetInt(args, 1))
	case cmdCarUpdName:
		return c.carUpdateAskName(ctx, pl.UserID, safeGetInt(args, 1))
	case cmdCarUpdYear:
		return c.carUpdateAskYear(ctx, pl.UserID, safeGetInt(args, 1))
	case cmdCarUpdPlate:
		return c.carUpdateAskPlate(ctx, pl.UserID, safeGetInt(args, 1))
	case cmdCarDel:
		return c.carDelAsk(ctx, pl.UserID, safeGetInt(args, 1))
	case cmdCarDelYes:
		return c.carDelYes(ctx, pl.UserID, safeGetInt(args, 1))
	case cmdCarFuel:
		return c.carShowFuelRecord(ctx, pl.UserID, safeGetInt(args, 1), 0)
	default:
		return c.carShowList(ctx, pl.UserID)
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
	html := fmt.Sprintf("🚘 <b>Name:</b> %s\n", car.Name)
	html += fmt.Sprintf("🏭 <b>Year:</b> %d\n", car.Year)
	if car.Plate != nil {
		html += fmt.Sprintf("🧾 <b>Plate:</b> %s\n", *car.Plate)
	} else {
		html += fmt.Sprintf("🧾 <b>Plate:</b> 🚫\n")
	}
	return html
}

func (c *CarCommand) carShowDetailsMenu(ctx context.Context, userID int64, carID int64) (Result, error) {
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
	res.AddKeyboardButton("« Back to my cars", c.Prefix())
	return res, nil
}

func (c *CarCommand) carShowList(ctx context.Context, userID int64) (Result, error) {
	cars, err := c.storage.SelectCarsFromDB(ctx, userID)
	if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}

	res := Result{Text: "Choose your car from the list below:"}
	for i, v := range cars {
		res.AddKeyboardButton(fmt.Sprintf("%s (%d)", v.Name, v.Year), commandf(c, cmdCarGet, v.ID))
		if (i+1)%2 == 0 {
			res.AddKeyboardRow()
		}
	}
	res.AddKeyboardRow()
	res.AddKeyboardButton("« Add new Car »", commandf(c, cmdCarAdd, nil))
	return res, nil
}

func (c *CarCommand) carAddStart(ctx context.Context, pl Payload) (Result, error) {
	c.draftCarInit(pl.UserID)
	return Result{Text: "Please choose a name for your car.", State: c.carAddName}, nil
}

func (c *CarCommand) carAddName(ctx context.Context, pl Payload) (Result, error) {
	c.draftCarSetName(pl.UserID, pl.Command)
	return Result{Text: "What is the model year?", State: c.carAddYear}, nil
}

func (c *CarCommand) carAddYear(ctx context.Context, pl Payload) (Result, error) {
	err := c.draftCarSetYear(pl.UserID, pl.Command)
	if err != nil {
		return Result{Text: "Please enter a valid number.", State: c.carAddYear}, nil
	}
	return Result{Text: "What is your plate number? /skip", State: c.carAddPlateAndSave}, nil
}

func (c *CarCommand) carAddPlateAndSave(ctx context.Context, pl Payload) (Result, error) {
	c.draftCarSetPlate(pl.UserID, pl.Command)
	carID, err := c.draftCarInsertIntoDB(ctx, pl.UserID)
	if err != nil {
		return Result{Text: "There is something wrong with our database, please try again."}, err
	}
	return c.carShowDetailsMenu(ctx, pl.UserID, carID)
}

func (c *CarCommand) carShowUpdateMenu(ctx context.Context, userID int64, carID int64) (Result, error) {
	car, err := c.storage.GetCarFromDB(ctx, userID, carID)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{Text: "Car not found."}, nil
	} else if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}

	res := Result{Text: c.formatCarDetails(car)}
	res.AddKeyboardButton("Edit Name", commandf(c, cmdCarUpdName, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Edit Year", commandf(c, cmdCarUpdYear, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Edit Plate", commandf(c, cmdCarUpdPlate, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, carID))
	return res, nil
}

func (c *CarCommand) carUpdateAskName(ctx context.Context, userID int64, carID int64) (Result, error) {
	if err := c.draftCarFetchFromDB(ctx, userID, carID); err != nil {
		return Result{Text: "Car not found."}, err
	}
	return Result{Text: "What is the new car name?", State: c.carUpdateSaveName}, nil
}

func (c *CarCommand) carUpdateAskYear(ctx context.Context, userID int64, carID int64) (Result, error) {
	if err := c.draftCarFetchFromDB(ctx, userID, carID); err != nil {
		return Result{Text: "Car not found."}, err
	}
	return Result{Text: "What is the new car year?", State: c.carUpdateSaveYear}, nil
}

func (c *CarCommand) carUpdateAskPlate(ctx context.Context, userID int64, carID int64) (Result, error) {
	if err := c.draftCarFetchFromDB(ctx, userID, carID); err != nil {
		return Result{Text: "Car not found."}, err
	}
	return Result{Text: "What is the new car plate? /skip", State: c.carUpdateSavePlate}, nil
}

func (c *CarCommand) carUpdateSaveName(ctx context.Context, pl Payload) (Result, error) {
	c.draftCarSetName(pl.UserID, pl.Command)
	if _, err := c.draftCarUpdateInDB(ctx, pl.UserID); err != nil {
		return Result{Text: "Update failed, try again."}, err
	}
	res := Result{Text: "Car name has been successfully updated!"}
	car := c.draftCars[pl.UserID]
	res.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, car.ID))
	return res, nil
}

func (c *CarCommand) carUpdateSaveYear(ctx context.Context, pl Payload) (Result, error) {
	err := c.draftCarSetYear(pl.UserID, pl.Command)
	if err != nil {
		return Result{Text: "Please enter a valid number.", State: c.carAddYear}, nil
	}
	if _, err := c.draftCarUpdateInDB(ctx, pl.UserID); err != nil {
		return Result{Text: "Update failed, try again."}, err
	}
	res := Result{Text: "Car year has been successfully updated!"}
	car := c.draftCars[pl.UserID]
	res.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, car.ID))
	return res, nil
}

func (c *CarCommand) carUpdateSavePlate(ctx context.Context, pl Payload) (Result, error) {
	c.draftCarSetPlate(pl.UserID, pl.Command)
	if _, err := c.draftCarUpdateInDB(ctx, pl.UserID); err != nil {
		return Result{Text: "Update failed, try again."}, err
	}
	res := Result{Text: "Car plate has been successfully updated!"}
	car := c.draftCars[pl.UserID]
	res.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, car.ID))
	return res, nil
}

func (c *CarCommand) carDelAsk(ctx context.Context, userID int64, carID int64) (Result, error) {
	car, err := c.storage.GetCarFromDB(ctx, userID, carID)
	if err != nil {
		return Result{Text: "Car not found."}, err
	}

	res := Result{Text: fmt.Sprintf("Are you sure you want to delete %s (%d)?", car.Name, car.Year)}
	res.AddKeyboardButton("Yes, delete the car", commandf(c, cmdCarDelYes, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("No", commandf(c, cmdCarGet, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Nope, nevermind", commandf(c, cmdCarGet, carID))
	return res, nil
}

func (c *CarCommand) carDelYes(ctx context.Context, userID int64, carID int64) (Result, error) {
	affected, err := c.storage.DeleteCarFromDB(ctx, userID, carID)
	if err != nil || affected != 1 {
		return Result{Text: "Car not found."}, err
	}

	res := Result{Text: "Car has been successfully deleted!"}
	res.AddKeyboardButton("« Back to my cars", c.Prefix())
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

func (c *CarCommand) formatFuelDetails() string {
	html := fmt.Sprintf("⛽ <b>Amount:</b> %s\n", "20L (Type 98)")
	html += fmt.Sprintf("💲 <b>Paid:</b> %s\n", "100 Eur (1Eur/L)")
	html += fmt.Sprintf("📍 <b>Traveled:</b> %s\n", "1,222 Km (1.1 L/Km)")
	html += fmt.Sprintf("📅 2023/01/01 11:11:11\n")
	return html
}

func (c *CarCommand) carShowFuelRecord(ctx context.Context, userID int64, carID int64, offset int) (Result, error) {
	// car, err := c.storage.GetCarFromDB(ctx, userID, carID)
	// if errors.Is(err, sql.ErrNoRows) {
	// 	return Result{Text: "Car not found."}, nil
	// } else if err != nil {
	// 	return Result{Text: "There is something wrong, please try again."}, err
	// }

	res := Result{Text: c.formatFuelDetails()}
	res.AddKeyboardButton("«5", commandf(c, cmdCarFuel, carID, offset-5))
	res.AddKeyboardButton("«1", commandf(c, cmdCarFuel, carID, offset-1))
	res.AddKeyboardButton("1/42", commandf(c, cmdCarFuel, carID, offset))
	res.AddKeyboardButton("1»", commandf(c, cmdCarFuel, carID, offset+1))
	res.AddKeyboardButton("5»", commandf(c, cmdCarFuel, carID, offset+5))
	res.AddKeyboardRow()
	res.AddKeyboardButton("« Delete »", commandf(c))
	res.AddKeyboardButton("« Add »", commandf(c))
	res.AddKeyboardRow()
	res.AddKeyboardButton("« Back to BMW (2022)", commandf(c, cmdCarGet, carID))
	return res, nil
}
