package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	st "mr-weasel/storage"
	"strconv"
	"strings"
	"time"
)

type CarCommand struct {
	storage   *st.CarStorage
	draftCars map[int64]*st.Car
	draftFuel map[int64]*st.Fuel
}

func NewCarCommand(storage *st.CarStorage) *CarCommand {
	c := &CarCommand{
		storage:   storage,
		draftCars: make(map[int64]*st.Car),
		draftFuel: make(map[int64]*st.Fuel),
	}
	return c
}

func (CarCommand) Prefix() string {
	return "/car"
}

func (CarCommand) Description() string {
	return "manage car costs"
}

const (
	cmdCarAdd        = "add"
	cmdCarGet        = "get"
	cmdCarUpd        = "upd"
	cmdCarUpdName    = "upd_name"
	cmdCarUpdYear    = "upd_year"
	cmdCarUpdPlate   = "upd_plate"
	cmdCarDel        = "del"
	cmdCarDelYes     = "del_yes"
	cmdCarFuelAdd    = "fuel_add"
	cmdCarFuelGet    = "fuel_get"
	cmdCarFuelDel    = "fuel_del"
	cmdCarFuelDelYes = "fuel_del_yes"
	cmdCarService    = "service"
)

func (c *CarCommand) Execute(ctx context.Context, pl Payload) (Result, error) {
	args := splitCommand(pl.Command, c.Prefix())
	switch safeGet(args, 0) {
	case cmdCarAdd:
		return c.addCarStart(ctx, pl)
	case cmdCarGet:
		return c.showCarDetails(ctx, pl.UserID, safeGetInt64(args, 1))
	case cmdCarUpd:
		return c.showCarUpdate(ctx, pl.UserID, safeGetInt64(args, 1))
	case cmdCarUpdName:
		return c.updateCarAskName(ctx, pl.UserID, safeGetInt64(args, 1))
	case cmdCarUpdYear:
		return c.updateCarAskYear(ctx, pl.UserID, safeGetInt64(args, 1))
	case cmdCarUpdPlate:
		return c.updateCarAskPlate(ctx, pl.UserID, safeGetInt64(args, 1))
	case cmdCarDel:
		return c.deleteCarAsk(ctx, pl.UserID, safeGetInt64(args, 1))
	case cmdCarDelYes:
		return c.deleteCarConfirm(ctx, pl.UserID, safeGetInt64(args, 1))
	case cmdCarFuelAdd:
		return c.addFuelStart(ctx, pl.UserID, safeGetInt64(args, 1))
	case cmdCarFuelGet:
		return c.showFuelDetails(ctx, pl.UserID, safeGetInt64(args, 1), 0)
	default:
		return c.showCarList(ctx, pl.UserID)
	}
}

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

func (c *CarCommand) showCarDetails(ctx context.Context, userID int64, carID int64) (Result, error) {
	car, err := c.storage.GetCarFromDB(ctx, userID, carID)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{Text: "Car not found."}, nil
	} else if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}

	res := Result{Text: c.formatCarDetails(car)}
	res.AddKeyboardButton("Fuel", commandf(c, cmdCarFuelGet, carID))
	res.AddKeyboardButton("Service", commandf(c, cmdCarService, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Edit Car", commandf(c, cmdCarUpd, carID))
	res.AddKeyboardButton("Delete Car", commandf(c, cmdCarDel, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("« Back to my cars", c.Prefix())
	return res, nil
}

func (c *CarCommand) showCarList(ctx context.Context, userID int64) (Result, error) {
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
	res.AddKeyboardButton("« New Car »", commandf(c, cmdCarAdd, nil))
	return res, nil
}

func (c *CarCommand) fetchDraftCarFromDB(ctx context.Context, userID int64, carID int64) error {
	car, err := c.storage.GetCarFromDB(ctx, userID, carID)
	if err == nil {
		c.draftCars[userID] = &car
	}
	return err
}

func (c *CarCommand) insertDraftCarIntoDB(ctx context.Context, userID int64) (int64, error) {
	return c.storage.InsertCarIntoDB(ctx, *c.draftCars[userID])
}

func (c *CarCommand) updateDraftCarInDB(ctx context.Context, userID int64) (int64, error) {
	return c.storage.UpdateCarInDB(ctx, *c.draftCars[userID])
}

func (c *CarCommand) newDraftCar(userID int64) {
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
	if input == "/skip" {
		c.draftCars[userID].Plate = nil
	} else {
		c.draftCars[userID].Plate = &input
	}
}

func (c *CarCommand) addCarStart(ctx context.Context, pl Payload) (Result, error) {
	c.newDraftCar(pl.UserID)
	return Result{Text: "Please choose a name for your car.", State: c.addCarName}, nil
}

func (c *CarCommand) addCarName(ctx context.Context, pl Payload) (Result, error) {
	c.setDraftCarName(pl.UserID, pl.Command)
	return Result{Text: "What is the model year?", State: c.addCarYear}, nil
}

func (c *CarCommand) addCarYear(ctx context.Context, pl Payload) (Result, error) {
	if err := c.setDraftCarYear(pl.UserID, pl.Command); err != nil {
		return Result{Text: "Please enter a valid number.", State: c.addCarYear}, nil
	}
	return Result{Text: "What is your plate number? /skip", State: c.addCarPlateAndSave}, nil
}

func (c *CarCommand) addCarPlateAndSave(ctx context.Context, pl Payload) (Result, error) {
	c.setDraftCarPlate(pl.UserID, pl.Command)
	carID, err := c.insertDraftCarIntoDB(ctx, pl.UserID)
	if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}
	return c.showCarDetails(ctx, pl.UserID, carID)
}

func (c *CarCommand) showCarUpdate(ctx context.Context, userID int64, carID int64) (Result, error) {
	car, err := c.storage.GetCarFromDB(ctx, userID, carID)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{Text: "Car not found."}, nil
	} else if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}

	res := Result{Text: c.formatCarDetails(car)}
	res.AddKeyboardButton("Set Name", commandf(c, cmdCarUpdName, carID))
	res.AddKeyboardButton("Set Year", commandf(c, cmdCarUpdYear, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Set Plate", commandf(c, cmdCarUpdPlate, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, carID))
	return res, nil
}

func (c *CarCommand) updateCarAskName(ctx context.Context, userID int64, carID int64) (Result, error) {
	if err := c.fetchDraftCarFromDB(ctx, userID, carID); err != nil {
		return Result{Text: "Car not found."}, err
	}
	return Result{Text: "What is the new car name?", State: c.updateCarSaveName}, nil
}

func (c *CarCommand) updateCarAskYear(ctx context.Context, userID int64, carID int64) (Result, error) {
	if err := c.fetchDraftCarFromDB(ctx, userID, carID); err != nil {
		return Result{Text: "Car not found."}, err
	}
	return Result{Text: "What is the new car year?", State: c.updateCarSaveYear}, nil
}

func (c *CarCommand) updateCarAskPlate(ctx context.Context, userID int64, carID int64) (Result, error) {
	if err := c.fetchDraftCarFromDB(ctx, userID, carID); err != nil {
		return Result{Text: "Car not found."}, err
	}
	return Result{Text: "What is the new car plate? /skip", State: c.updateCarSavePlate}, nil
}

func (c *CarCommand) updateCarSaveName(ctx context.Context, pl Payload) (Result, error) {
	c.setDraftCarName(pl.UserID, pl.Command)
	if _, err := c.updateDraftCarInDB(ctx, pl.UserID); err != nil {
		return Result{Text: "Update failed, try again."}, err
	}
	res := Result{Text: "Car name has been successfully updated!"}
	car := c.draftCars[pl.UserID]
	res.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, car.ID))
	return res, nil
}

func (c *CarCommand) updateCarSaveYear(ctx context.Context, pl Payload) (Result, error) {
	if err := c.setDraftCarYear(pl.UserID, pl.Command); err != nil {
		return Result{Text: "Please enter a valid number.", State: c.updateCarSaveYear}, nil
	}
	if _, err := c.updateDraftCarInDB(ctx, pl.UserID); err != nil {
		return Result{Text: "Update failed, try again."}, err
	}
	res := Result{Text: "Car year has been successfully updated!"}
	car := c.draftCars[pl.UserID]
	res.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, car.ID))
	return res, nil
}

func (c *CarCommand) updateCarSavePlate(ctx context.Context, pl Payload) (Result, error) {
	c.setDraftCarPlate(pl.UserID, pl.Command)
	if _, err := c.updateDraftCarInDB(ctx, pl.UserID); err != nil {
		return Result{Text: "Update failed, try again."}, err
	}
	res := Result{Text: "Car plate has been successfully updated!"}
	car := c.draftCars[pl.UserID]
	res.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, car.ID))
	return res, nil
}

func (c *CarCommand) deleteCarAsk(ctx context.Context, userID int64, carID int64) (Result, error) {
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

func (c *CarCommand) deleteCarConfirm(ctx context.Context, userID int64, carID int64) (Result, error) {
	affected, err := c.storage.DeleteCarFromDB(ctx, userID, carID)
	if err != nil || affected != 1 {
		return Result{Text: "Car not found."}, err
	}
	res := Result{Text: "Car has been successfully deleted!"}
	res.AddKeyboardButton("« Back to my cars", c.Prefix())
	return res, nil

}

func (c *CarCommand) formatFuelDetails(fuel st.Fuel) string {
	html := fmt.Sprintf("⛽ <b>Liters:</b> %.2fLi (%s)\n", fuel.GetLiters(), fuel.Type)
	html += fmt.Sprintf("💲 <b>Paid:</b> %.2f€ (%.2f€/Li)\n", fuel.GetEuro(), fuel.GetEurPerLiter())
	html += fmt.Sprintf("📍 <b>Traveled:</b> %d (%.2fL/Km)\n", fuel.KilometersR, fuel.GetLitersPerKilometer())
	html += fmt.Sprintf("📅 %s\n", fuel.GetTimestamp().UTC().Format("Monday, 02 January 2006"))
	return html
}

func (c *CarCommand) showFuelDetails(ctx context.Context, userID int64, carID int64, offset int) (Result, error) {
	res := Result{}
	fuel, err := c.storage.GetFuelFromDB(ctx, userID, carID, offset)
	if errors.Is(err, sql.ErrNoRows) {
		res.Text = "No fuel records."
	} else if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	} else {
		res.Text = c.formatFuelDetails(fuel)
	}

	res.AddKeyboardButton("«5", commandf(c, cmdCarFuelGet, carID, offset-5))
	res.AddKeyboardButton("«1", commandf(c, cmdCarFuelGet, carID, offset-1))
	res.AddKeyboardButton("1/42", commandf(c, cmdCarFuelGet, carID, offset))
	res.AddKeyboardButton("1»", commandf(c, cmdCarFuelGet, carID, offset+1))
	res.AddKeyboardButton("5»", commandf(c, cmdCarFuelGet, carID, offset+5))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Delete", commandf(c))
	res.AddKeyboardButton("Add", commandf(c, cmdCarFuelAdd, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("« Back to BMW (2022)", commandf(c, cmdCarGet, carID))
	return res, nil
}

func (c *CarCommand) insertDraftFuelIntoDB(ctx context.Context, userID int64) (int64, error) {
	return c.storage.InsertFuelIntoDB(ctx, *c.draftFuel[userID])
}

func (c *CarCommand) newDraftFuel(userID int64, carID int64) {
	c.draftFuel[userID] = &st.Fuel{CarID: carID}
}

func (c *CarCommand) setDraftFuelTimestamp(userID int64, input string) error {
	timestamp, err := strconv.Atoi(input)
	c.draftFuel[userID].Timestamp = int64(timestamp)
	return err
}

func (c *CarCommand) setDraftFuelType(userID int64, input string) {
	c.draftFuel[userID].Type = input
}

func (c *CarCommand) setDraftFuelLiters(userID int64, input string) error {
	liters, err := strconv.ParseFloat(input, 64)
	c.draftFuel[userID].Milliliters = int64(liters * 1000)
	return err
}

func (c *CarCommand) setDraftFuelKilometers(userID int64, input string) error {
	kilometers, err := strconv.Atoi(input)
	c.draftFuel[userID].Kilometers = int64(kilometers)
	return err
}

func (c *CarCommand) setDraftFuelEuros(userID int64, input string) error {
	euro, err := strconv.ParseFloat(input, 64)
	c.draftFuel[userID].Cents = int64(euro * 100)
	return err
}

func (c *CarCommand) addFuelStart(ctx context.Context, userID int64, carID int64) (Result, error) {
	c.newDraftFuel(userID, carID)
	res := Result{Text: "Please pick a receipt date.", State: c.addFuelTimestamp}
	res.AddKeyboardCalendar(time.Now().Year(), time.Now().Month())
	return res, nil
}

func (c *CarCommand) addFuelTimestamp(ctx context.Context, pl Payload) (Result, error) {
	if s := strings.Split(pl.Command, " "); len(s) == 2 {
		res := Result{Text: "Please pick a receipt date.", State: c.addFuelTimestamp}
		res.AddKeyboardCalendar(safeGetInt(s, 0), time.Month(safeGetInt(s, 1)))
		return res, nil
	}
	if err := c.setDraftFuelTimestamp(pl.UserID, pl.Command); err != nil {
		return Result{State: c.addFuelTimestamp}, nil
	}
	res := Result{Text: "What is the fuel type?", State: c.addFuelType}
	res.AddKeyboardRow() // remove calendar keyboard
	return res, nil
}

func (c *CarCommand) addFuelType(ctx context.Context, pl Payload) (Result, error) {
	c.setDraftFuelType(pl.UserID, pl.Command)
	return Result{Text: "What is the fuel amount in Liters?", State: c.addFuelLiters}, nil
}

func (c *CarCommand) addFuelLiters(ctx context.Context, pl Payload) (Result, error) {
	if err := c.setDraftFuelLiters(pl.UserID, pl.Command); err != nil {
		return Result{Text: "Please enter a valid decimal number.", State: c.addFuelLiters}, nil
	}
	return Result{Text: "What is your total mileage now in Kilometers?", State: c.addFuelKilometers}, nil
}

func (c *CarCommand) addFuelKilometers(ctx context.Context, pl Payload) (Result, error) {
	if err := c.setDraftFuelKilometers(pl.UserID, pl.Command); err != nil {
		return Result{Text: "Please enter a valid whole number.", State: c.addFuelKilometers}, nil
	}
	return Result{Text: "How much money did you spend in Euros?", State: c.addFuelEurosAndSave}, nil
}

func (c *CarCommand) addFuelEurosAndSave(ctx context.Context, pl Payload) (Result, error) {
	if err := c.setDraftFuelEuros(pl.UserID, pl.Command); err != nil {
		return Result{Text: "Please enter a valid decimal number.", State: c.addFuelEurosAndSave}, nil
	}
	if _, err := c.insertDraftFuelIntoDB(ctx, pl.UserID); err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}
	return c.showFuelDetails(ctx, pl.UserID, c.draftFuel[pl.UserID].CarID, 0)
}
