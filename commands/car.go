package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	st "mr-weasel/storage"
)

type CarCommand struct {
	storage      *st.CarStorage
	draftCars    map[int64]*st.CarBase
	draftFuel    map[int64]*st.FuelBase
	draftService map[int64]*st.ServiceBase
	draftLease   map[int64]*st.LeaseBase
}

func NewCarCommand(storage *st.CarStorage) *CarCommand {
	c := &CarCommand{
		storage:      storage,
		draftCars:    make(map[int64]*st.CarBase),
		draftFuel:    make(map[int64]*st.FuelBase),
		draftService: make(map[int64]*st.ServiceBase),
		draftLease:   make(map[int64]*st.LeaseBase),
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
	cmdCarAdd           = "add"
	cmdCarGet           = "get"
	cmdCarUpd           = "upd"
	cmdCarUpdName       = "upd_name"
	cmdCarUpdYear       = "upd_year"
	cmdCarUpdPlate      = "upd_plate"
	cmdCarDelAsk        = "del"
	cmdCarDelYes        = "del_yes"
	cmdCarFuelAdd       = "fuel_add"
	cmdCarFuelGet       = "fuel_get"
	cmdCarFuelDelAsk    = "fuel_del"
	cmdCarFuelDelYes    = "fuel_del_yes"
	cmdCarServiceAdd    = "service_add"
	cmdCarServiceGet    = "service_get"
	cmdCarServiceDelAsk = "service_del"
	cmdCarServiceDelYes = "service_del_yes"
	cmdCarLeaseAdd      = "lease_add"
	cmdCarLeaseGet      = "lease_get"
	cmdCarLeaseDelAsk   = "lease_del"
	cmdCarLeaseDelYes   = "lease_del_yes"
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
	case cmdCarDelAsk:
		return c.deleteCarAsk(ctx, pl.UserID, safeGetInt64(args, 1))
	case cmdCarDelYes:
		return c.deleteCarConfirm(ctx, pl.UserID, safeGetInt64(args, 1))
	case cmdCarFuelAdd:
		return c.addFuelStart(ctx, pl.UserID, safeGetInt64(args, 1))
	case cmdCarFuelGet:
		return c.showFuelDetails(ctx, pl.UserID, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarFuelDelAsk:
		return c.deleteFuelAsk(ctx, pl.UserID, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarFuelDelYes:
		return c.deleteFuelConfirm(ctx, pl.UserID, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarServiceAdd:
		return c.addServiceStart(ctx, pl.UserID, safeGetInt64(args, 1))
	case cmdCarServiceGet:
		return c.showServiceDetails(ctx, pl.UserID, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarServiceDelAsk:
		return c.deleteServiceAsk(ctx, pl.UserID, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarServiceDelYes:
		return c.deleteServiceConfirm(ctx, pl.UserID, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarLeaseAdd:
		return c.addLeaseStart(ctx, pl.UserID, safeGetInt64(args, 1))
	case cmdCarLeaseGet:
		return c.showLeaseDetails(ctx, pl.UserID, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarLeaseDelAsk:
		return c.deleteLeaseAsk(ctx, pl.UserID, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarLeaseDelYes:
		return c.deleteLeaseConfirm(ctx, pl.UserID, safeGetInt64(args, 1), safeGetInt64(args, 2))
	default:
		return c.showCarList(ctx, pl.UserID)
	}
}

func (c *CarCommand) formatCarDetails(car st.CarDetails) string {
	html := fmt.Sprintf("🚘 <b>Car:</b> %s (%d)\n", car.Name, car.Year)
	if car.Plate.Valid {
		html += fmt.Sprintf("🧾 <b>Licence Plate:</b> %s\n", car.Plate.String)
	} else {
		html += fmt.Sprintf("🧾 <b>Licence Plate:</b> 🚫\n")
	}
	html += fmt.Sprintf("📍 <b>Mileage:</b> %dKm\n", car.Kilometers)
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
	res.AddKeyboardButton("Service", commandf(c, cmdCarServiceGet, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Edit Car", commandf(c, cmdCarUpd, carID))
	res.AddKeyboardButton("Delete Car", commandf(c, cmdCarDelAsk, carID))
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
		c.draftCars[userID] = &car.CarBase
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
	c.draftCars[userID] = &st.CarBase{UserID: userID}
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
		c.draftCars[userID].Plate.Valid = false
	} else {
		c.draftCars[userID].Plate.Valid = true
		c.draftCars[userID].Plate.String = input
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

func (c *CarCommand) formatFuelDetails(fuel st.FuelDetails) string {
	html := fmt.Sprintf("⛽ <b>Liters:</b> %.2fL (%s)\n", fuel.GetLiters(), fuel.Type)
	html += fmt.Sprintf("💲 <b>Paid:</b> %.2f€ (%.2fEur/L)\n", fuel.GetEuro(), fuel.GetEurPerLiter())
	html += fmt.Sprintf("📍 <b>Traveled:</b> %dKm (%.2fL/Km)\n", fuel.KilometersR, fuel.GetLitersPerKilometer())
	html += fmt.Sprintf("🏭 <b>Total:</b> %dKm\n", fuel.Kilometers)
	html += fmt.Sprintf("📅 %s\n", fuel.GetTimestamp())
	return html
}

func (c *CarCommand) showFuelDetails(ctx context.Context, userID int64, carID int64, offset int64) (Result, error) {
	res := Result{}
	fuel, err := c.storage.GetFuelFromDB(ctx, userID, carID, offset)
	if errors.Is(err, sql.ErrNoRows) {
		res.Text = "No fuel receipts found."
	} else if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	} else {
		res.Text = c.formatFuelDetails(fuel)
		res.AddKeyboardPagination(offset, fuel.CountRows, commandf(c, cmdCarFuelGet, carID))
		res.AddKeyboardRow()
		res.AddKeyboardButton("Delete", commandf(c, cmdCarFuelDelAsk, carID, fuel.ID))
	}
	res.AddKeyboardButton("Add", commandf(c, cmdCarFuelAdd, carID))
	res.AddKeyboardRow()
	car, _ := c.storage.GetCarFromDB(ctx, userID, carID)
	res.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, carID))
	return res, nil
}

func (c *CarCommand) insertDraftFuelIntoDB(ctx context.Context, userID int64) (int64, error) {
	return c.storage.InsertFuelIntoDB(ctx, *c.draftFuel[userID])
}

func (c *CarCommand) newDraftFuel(userID int64, carID int64) {
	c.draftFuel[userID] = &st.FuelBase{CarID: carID}
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
	res := Result{}
	if res.UpdateKeyboardCalendar(pl.Command) {
		return res, nil
	} else if c.setDraftFuelTimestamp(pl.UserID, pl.Command) != nil {
		return res, nil
	}

	res.Text, res.State = "What is the fuel type?", c.addFuelType
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

func (c *CarCommand) deleteFuelAsk(ctx context.Context, userID int64, carID int64, fuelID int64) (Result, error) {
	res := Result{Text: "Are you sure you want to delete the selected receipt?"}
	res.AddKeyboardButton("Yes, delete the receipt", commandf(c, cmdCarFuelDelYes, carID, fuelID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("No", commandf(c, cmdCarFuelGet, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Nope, nevermind", commandf(c, cmdCarFuelGet, carID))
	return res, nil
}

func (c *CarCommand) deleteFuelConfirm(ctx context.Context, userID int64, carID int64, fuelID int64) (Result, error) {
	affected, err := c.storage.DeleteFuelFromDB(ctx, userID, fuelID)
	if err != nil || affected != 1 {
		return Result{Text: "Receipt not found."}, err
	}
	res := Result{Text: "Receipt has been successfully deleted!"}
	res.AddKeyboardButton("« Back to my receipts", commandf(c, cmdCarFuelGet, carID))
	return res, nil
}

func (c *CarCommand) formatServiceDetails(service st.ServiceDetails) string {
	html := fmt.Sprintf("🛠️ %s\n", service.Description)
	html += fmt.Sprintf("💲 <b>Paid:</b> %.2f€\n", service.GetEuro())
	html += fmt.Sprintf("📅 %s\n", service.GetTimestamp())
	return html
}

func (c *CarCommand) showServiceDetails(ctx context.Context, userID int64, carID int64, offset int64) (Result, error) {
	res := Result{}
	service, err := c.storage.GetServiceFromDB(ctx, userID, carID, offset)
	if errors.Is(err, sql.ErrNoRows) {
		res.Text = "No service receipts found."
	} else if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	} else {
		res.Text = c.formatServiceDetails(service)
		res.AddKeyboardPagination(offset, service.CountRows, commandf(c, cmdCarServiceGet, carID))
		res.AddKeyboardRow()
		res.AddKeyboardButton("Delete", commandf(c, cmdCarServiceDelAsk, carID, service.ID))
	}
	res.AddKeyboardButton("Add", commandf(c, cmdCarServiceAdd, carID))
	res.AddKeyboardRow()
	car, _ := c.storage.GetCarFromDB(ctx, userID, carID)
	res.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, carID))
	return res, nil
}

func (c *CarCommand) insertDraftServiceIntoDB(ctx context.Context, userID int64) (int64, error) {
	return c.storage.InsertServiceIntoDB(ctx, *c.draftService[userID])
}

func (c *CarCommand) newDraftService(userID int64, carID int64) {
	c.draftService[userID] = &st.ServiceBase{CarID: carID}
}

func (c *CarCommand) setDraftServiceTimestamp(userID int64, input string) error {
	timestamp, err := strconv.Atoi(input)
	c.draftService[userID].Timestamp = int64(timestamp)
	return err
}

func (c *CarCommand) setDraftServiceDescription(userID int64, input string) {
	c.draftService[userID].Description = input
}

func (c *CarCommand) setDraftServiceEuros(userID int64, input string) error {
	euro, err := strconv.ParseFloat(input, 64)
	c.draftService[userID].Cents = int64(euro * 100)
	return err
}

func (c *CarCommand) addServiceStart(ctx context.Context, userID int64, carID int64) (Result, error) {
	c.newDraftService(userID, carID)
	res := Result{Text: "Please pick a receipt date.", State: c.addServiceTimestamp}
	res.AddKeyboardCalendar(time.Now().Year(), time.Now().Month())
	return res, nil
}

func (c *CarCommand) addServiceTimestamp(ctx context.Context, pl Payload) (Result, error) {
	res := Result{}
	if res.UpdateKeyboardCalendar(pl.Command) {
		return res, nil
	} else if c.setDraftServiceTimestamp(pl.UserID, pl.Command) != nil {
		return res, nil
	}

	res.Text, res.State = "Provide service description.", c.addServiceDescription
	res.AddKeyboardRow() // remove calendar keyboard
	return res, nil
}

func (c *CarCommand) addServiceDescription(ctx context.Context, pl Payload) (Result, error) {
	c.setDraftServiceDescription(pl.UserID, pl.Command)
	return Result{Text: "How much money did you spend in Euros?", State: c.addServiceEurosAndSave}, nil
}

func (c *CarCommand) addServiceEurosAndSave(ctx context.Context, pl Payload) (Result, error) {
	if err := c.setDraftServiceEuros(pl.UserID, pl.Command); err != nil {
		return Result{Text: "Please enter a valid decimal number.", State: c.addServiceEurosAndSave}, nil
	}
	if _, err := c.insertDraftServiceIntoDB(ctx, pl.UserID); err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}
	return c.showServiceDetails(ctx, pl.UserID, c.draftService[pl.UserID].CarID, 0)
}

func (c *CarCommand) deleteServiceAsk(ctx context.Context, userID int64, carID int64, serviceID int64) (Result, error) {
	res := Result{Text: "Are you sure you want to delete the selected receipt?"}
	res.AddKeyboardButton("Yes, delete the receipt", commandf(c, cmdCarServiceDelYes, carID, serviceID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("No", commandf(c, cmdCarServiceGet, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Nope, nevermind", commandf(c, cmdCarServiceGet, carID))
	return res, nil
}

func (c *CarCommand) deleteServiceConfirm(ctx context.Context, userID int64, carID int64, serviceID int64) (Result, error) {
	affected, err := c.storage.DeleteServiceFromDB(ctx, userID, serviceID)
	if err != nil || affected != 1 {
		return Result{Text: "Receipt not found."}, err
	}
	res := Result{Text: "Receipt has been successfully deleted!"}
	res.AddKeyboardButton("« Back to my receipts", commandf(c, cmdCarServiceGet, carID))
	return res, nil
}

func (c *CarCommand) formatLeaseDetails(lease st.LeaseDetails) string {
	html := fmt.Sprintf("🛠️ %s\n", lease.Description)
	html += fmt.Sprintf("💲 <b>Paid:</b> %.2f€\n", lease.GetEuro())
	html += fmt.Sprintf("📅 %s\n", lease.GetTimestamp())
	return html
}

func (c *CarCommand) showLeaseDetails(ctx context.Context, userID int64, carID int64, offset int64) (Result, error) {
	res := Result{}
	lease, err := c.storage.GetLeaseFromDB(ctx, userID, carID, offset)
	if errors.Is(err, sql.ErrNoRows) {
		res.Text = "No lease receipts found."
	} else if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	} else {
		res.Text = c.formatLeaseDetails(lease)
		res.AddKeyboardPagination(offset, lease.CountRows, commandf(c, cmdCarLeaseGet, carID))
		res.AddKeyboardRow()
		res.AddKeyboardButton("Delete", commandf(c, cmdCarLeaseDelAsk, carID, lease.ID))
	}
	res.AddKeyboardButton("Add", commandf(c, cmdCarLeaseAdd, carID))
	res.AddKeyboardRow()
	car, _ := c.storage.GetCarFromDB(ctx, userID, carID)
	res.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, carID))
	return res, nil
}

func (c *CarCommand) insertDraftLeaseIntoDB(ctx context.Context, userID int64) (int64, error) {
	return c.storage.InsertLeaseIntoDB(ctx, *c.draftLease[userID])
}

func (c *CarCommand) newDraftLease(userID int64, carID int64) {
	c.draftLease[userID] = &st.LeaseBase{CarID: carID}
}

func (c *CarCommand) setDraftLeaseTimestamp(userID int64, input string) error {
	timestamp, err := strconv.Atoi(input)
	c.draftLease[userID].Timestamp = int64(timestamp)
	return err
}

func (c *CarCommand) setDraftLeaseDescription(userID int64, input string) {
	c.draftLease[userID].Description = input
}

func (c *CarCommand) setDraftLeaseEuros(userID int64, input string) error {
	euro, err := strconv.ParseFloat(input, 64)
	c.draftLease[userID].Cents = int64(euro * 100)
	return err
}

func (c *CarCommand) addLeaseStart(ctx context.Context, userID int64, carID int64) (Result, error) {
	c.newDraftLease(userID, carID)
	res := Result{Text: "Please pick a receipt date.", State: c.addLeaseTimestamp}
	res.AddKeyboardCalendar(time.Now().Year(), time.Now().Month())
	return res, nil
}

func (c *CarCommand) addLeaseTimestamp(ctx context.Context, pl Payload) (Result, error) {
	res := Result{}
	if res.UpdateKeyboardCalendar(pl.Command) {
		return res, nil
	} else if c.setDraftLeaseTimestamp(pl.UserID, pl.Command) != nil {
		return res, nil
	}

	res.Text, res.State = "Provide lease description.", c.addLeaseDescription
	res.AddKeyboardRow() // remove calendar keyboard
	return res, nil
}

func (c *CarCommand) addLeaseDescription(ctx context.Context, pl Payload) (Result, error) {
	c.setDraftLeaseDescription(pl.UserID, pl.Command)
	return Result{Text: "How much money did you spend in Euros?", State: c.addLeaseEurosAndSave}, nil
}

func (c *CarCommand) addLeaseEurosAndSave(ctx context.Context, pl Payload) (Result, error) {
	if err := c.setDraftLeaseEuros(pl.UserID, pl.Command); err != nil {
		return Result{Text: "Please enter a valid decimal number.", State: c.addLeaseEurosAndSave}, nil
	}
	if _, err := c.insertDraftLeaseIntoDB(ctx, pl.UserID); err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}
	return c.showLeaseDetails(ctx, pl.UserID, c.draftLease[pl.UserID].CarID, 0)
}

func (c *CarCommand) deleteLeaseAsk(ctx context.Context, userID int64, carID int64, leaseID int64) (Result, error) {
	res := Result{Text: "Are you sure you want to delete the selected receipt?"}
	res.AddKeyboardButton("Yes, delete the receipt", commandf(c, cmdCarLeaseDelYes, carID, leaseID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("No", commandf(c, cmdCarLeaseGet, carID))
	res.AddKeyboardRow()
	res.AddKeyboardButton("Nope, nevermind", commandf(c, cmdCarLeaseGet, carID))
	return res, nil
}

func (c *CarCommand) deleteLeaseConfirm(ctx context.Context, userID int64, carID int64, leaseID int64) (Result, error) {
	affected, err := c.storage.DeleteLeaseFromDB(ctx, userID, leaseID)
	if err != nil || affected != 1 {
		return Result{Text: "Receipt not found."}, err
	}
	res := Result{Text: "Receipt has been successfully deleted!"}
	res.AddKeyboardButton("« Back to my receipts", commandf(c, cmdCarLeaseGet, carID))
	return res, nil
}
