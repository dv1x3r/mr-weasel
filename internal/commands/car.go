package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	st "mr-weasel/internal/storage"
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
	return "manage car expenses"
}

const (
	cmdCarAdd           = "add"
	cmdCarGet           = "get"
	cmdCarUpd           = "upd"
	cmdCarUpdName       = "upd_name"
	cmdCarUpdYear       = "upd_year"
	cmdCarUpdPlate      = "upd_plate"
	cmdCarUpdPrice      = "upd_price"
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

func (c *CarCommand) Execute(ctx context.Context, pl Payload) {
	args := splitCommand(pl.Command, c.Prefix())
	switch safeGet(args, 0) {
	case cmdCarAdd:
		c.addCarStart(ctx, pl)
	case cmdCarGet:
		c.showCarDetails(ctx, pl, safeGetInt64(args, 1))
	case cmdCarUpd:
		c.showCarUpdate(ctx, pl, safeGetInt64(args, 1))
	case cmdCarUpdName:
		c.updateCarAskName(ctx, pl, safeGetInt64(args, 1))
	case cmdCarUpdYear:
		c.updateCarAskYear(ctx, pl, safeGetInt64(args, 1))
	case cmdCarUpdPlate:
		c.updateCarAskPlate(ctx, pl, safeGetInt64(args, 1))
	case cmdCarUpdPrice:
		c.updateCarAskPrice(ctx, pl, safeGetInt64(args, 1))
	case cmdCarDelAsk:
		c.deleteCarAsk(ctx, pl, safeGetInt64(args, 1))
	case cmdCarDelYes:
		c.deleteCarConfirm(ctx, pl, safeGetInt64(args, 1))
	case cmdCarFuelAdd:
		c.addFuelStart(ctx, pl, safeGetInt64(args, 1))
	case cmdCarFuelGet:
		c.showFuelDetails(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarFuelDelAsk:
		c.deleteFuelAsk(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarFuelDelYes:
		c.deleteFuelConfirm(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarServiceAdd:
		c.addServiceStart(ctx, pl, safeGetInt64(args, 1))
	case cmdCarServiceGet:
		c.showServiceDetails(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarServiceDelAsk:
		c.deleteServiceAsk(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarServiceDelYes:
		c.deleteServiceConfirm(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarLeaseAdd:
		c.addLeaseStart(ctx, pl, safeGetInt64(args, 1))
	case cmdCarLeaseGet:
		c.showLeaseDetails(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarLeaseDelAsk:
		c.deleteLeaseAsk(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	case cmdCarLeaseDelYes:
		c.deleteLeaseConfirm(ctx, pl, safeGetInt64(args, 1), safeGetInt64(args, 2))
	default:
		c.showCarList(ctx, pl)
	}
}

func (c *CarCommand) formatCarDetails(car st.CarDetails) string {
	str := fmt.Sprintf("🚘 <b>Car:</b> %s (%d)\n", _es(car.Name), car.Year)
	if car.Price.Valid {
		str += fmt.Sprintf("💲 <b>Price:</b> %d€\n", car.Price.Int64)
	} else {
		str += fmt.Sprintf("💲 <b>Price:</b> 🚫\n")
	}
	str += fmt.Sprintf("📍 <b>Mileage:</b> %dKm\n", car.Kilometers)
	if car.Plate.Valid {
		str += fmt.Sprintf("🧾 <b>Licence Plate:</b> %s\n", _es(car.Plate.String))
	} else {
		str += fmt.Sprintf("🧾 <b>Licence Plate:</b> 🚫\n")
	}
	return str
}

func (c *CarCommand) showCarDetails(ctx context.Context, pl Payload, carID int64) {
	res := Result{}
	car, err := c.storage.GetCarFromDB(ctx, pl.UserID, carID)
	if errors.Is(err, sql.ErrNoRows) {
		res.Text = "Car not found."
	} else if err != nil {
		res.Text, res.Error = "There is something wrong, please try again.", err
	} else {
		res.Text = c.formatCarDetails(car)
		res.InlineMarkup.AddKeyboardButton("Fuel", commandf(c, cmdCarFuelGet, carID))
		res.InlineMarkup.AddKeyboardButton("Service", commandf(c, cmdCarServiceGet, carID))
		res.InlineMarkup.AddKeyboardButton("Lease", commandf(c, cmdCarLeaseGet, carID))
		res.InlineMarkup.AddKeyboardRow()
		res.InlineMarkup.AddKeyboardButton("Edit Car", commandf(c, cmdCarUpd, carID))
		res.InlineMarkup.AddKeyboardRow()
	}
	res.InlineMarkup.AddKeyboardButton("« Back to my cars", c.Prefix())
	pl.ResultChan <- res
}

func (c *CarCommand) showCarList(ctx context.Context, pl Payload) {
	cars, err := c.storage.SelectCarsFromDB(ctx, pl.UserID)
	if err != nil {
		pl.ResultChan <- Result{Text: "There is something wrong, please try again.", Error: err}
		return
	}

	res := Result{Text: "Choose your car from the list below:"}
	for i, v := range cars {
		res.InlineMarkup.AddKeyboardButton(fmt.Sprintf("%s (%d)", v.Name, v.Year), commandf(c, cmdCarGet, v.ID))
		if (i+1)%2 == 0 {
			res.InlineMarkup.AddKeyboardRow()
		}
	}
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("« New Car »", commandf(c, cmdCarAdd, nil))
	pl.ResultChan <- res
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

func (c *CarCommand) setDraftCarPrice(userID int64, input string) error {
	if input == "/skip" {
		c.draftCars[userID].Price.Valid = false
		return nil
	} else {
		price, err := strconv.Atoi(input)
		c.draftCars[userID].Price.Int64 = int64(price)
		c.draftCars[userID].Price.Valid = true
		return err
	}
}

func (c *CarCommand) addCarStart(ctx context.Context, pl Payload) {
	c.newDraftCar(pl.UserID)
	pl.ResultChan <- Result{Text: "Please choose a name for your car.", State: c.addCarName}
}

func (c *CarCommand) addCarName(ctx context.Context, pl Payload) {
	c.setDraftCarName(pl.UserID, pl.Command)
	pl.ResultChan <- Result{Text: "What is the model year?", State: c.addCarYear}
}

func (c *CarCommand) addCarYear(ctx context.Context, pl Payload) {
	if err := c.setDraftCarYear(pl.UserID, pl.Command); err != nil {
		pl.ResultChan <- Result{Text: "Please enter a valid number.", State: c.addCarYear}
	} else {
		pl.ResultChan <- Result{Text: "What is your plate number? /skip", State: c.addCarPlate}
	}
}

func (c *CarCommand) addCarPlate(ctx context.Context, pl Payload) {
	c.setDraftCarPlate(pl.UserID, pl.Command)
	pl.ResultChan <- Result{Text: "What is the price? /skip", State: c.addCarPriceAndSave}
}

func (c *CarCommand) addCarPriceAndSave(ctx context.Context, pl Payload) {
	if err := c.setDraftCarPrice(pl.UserID, pl.Command); err != nil {
		pl.ResultChan <- Result{Text: "Please enter a valid number.", State: c.addCarPriceAndSave}
		return
	}
	carID, err := c.insertDraftCarIntoDB(ctx, pl.UserID)
	if err != nil {
		pl.ResultChan <- Result{Text: "There is something wrong, please try again.", Error: err}
		return
	}
	c.showCarDetails(ctx, pl, carID)
}

func (c *CarCommand) showCarUpdate(ctx context.Context, pl Payload, carID int64) {
	res := Result{}
	car, err := c.storage.GetCarFromDB(ctx, pl.UserID, carID)
	if errors.Is(err, sql.ErrNoRows) {
		res.Text = "Car not found."
	} else if err != nil {
		res.Text, res.Error = "There is something wrong, please try again.", err
	} else {
		res.Text = c.formatCarDetails(car)
		res.InlineMarkup.AddKeyboardButton("Set Name", commandf(c, cmdCarUpdName, carID))
		res.InlineMarkup.AddKeyboardButton("Set Year", commandf(c, cmdCarUpdYear, carID))
		res.InlineMarkup.AddKeyboardRow()
		res.InlineMarkup.AddKeyboardButton("Set Plate", commandf(c, cmdCarUpdPlate, carID))
		res.InlineMarkup.AddKeyboardButton("Set Price", commandf(c, cmdCarUpdPrice, carID))
		res.InlineMarkup.AddKeyboardRow()
		res.InlineMarkup.AddKeyboardButton("Delete Car", commandf(c, cmdCarDelAsk, carID))
		res.InlineMarkup.AddKeyboardRow()
	}
	res.InlineMarkup.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, carID))
	pl.ResultChan <- res
}

func (c *CarCommand) updateCarAskName(ctx context.Context, pl Payload, carID int64) {
	if err := c.fetchDraftCarFromDB(ctx, pl.UserID, carID); err != nil {
		pl.ResultChan <- Result{Text: "Car not found.", Error: err}
	} else {
		pl.ResultChan <- Result{Text: "What is the new car name?", State: c.updateCarSaveName}
	}
}

func (c *CarCommand) updateCarAskYear(ctx context.Context, pl Payload, carID int64) {
	if err := c.fetchDraftCarFromDB(ctx, pl.UserID, carID); err != nil {
		pl.ResultChan <- Result{Text: "Car not found.", Error: err}
	} else {
		pl.ResultChan <- Result{Text: "What is the new car year?", State: c.updateCarSaveYear}
	}
}

func (c *CarCommand) updateCarAskPlate(ctx context.Context, pl Payload, carID int64) {
	if err := c.fetchDraftCarFromDB(ctx, pl.UserID, carID); err != nil {
		pl.ResultChan <- Result{Text: "Car not found.", Error: err}
	} else {
		pl.ResultChan <- Result{Text: "What is the new car plate? /skip", State: c.updateCarSavePlate}
	}
}

func (c *CarCommand) updateCarAskPrice(ctx context.Context, pl Payload, carID int64) {
	if err := c.fetchDraftCarFromDB(ctx, pl.UserID, carID); err != nil {
		pl.ResultChan <- Result{Text: "Car not found.", Error: err}
	} else {
		pl.ResultChan <- Result{Text: "What is the new car price? /skip", State: c.updateCarSavePrice}
	}
}

func (c *CarCommand) updateCarSaveName(ctx context.Context, pl Payload) {
	c.setDraftCarName(pl.UserID, pl.Command)
	if _, err := c.updateDraftCarInDB(ctx, pl.UserID); err != nil {
		pl.ResultChan <- Result{Text: "Update failed, try again.", Error: err}
		return
	}
	res := Result{Text: "Car name has been successfully updated!"}
	car := c.draftCars[pl.UserID]
	res.InlineMarkup.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, car.ID))
	pl.ResultChan <- res
}

func (c *CarCommand) updateCarSaveYear(ctx context.Context, pl Payload) {
	if err := c.setDraftCarYear(pl.UserID, pl.Command); err != nil {
		pl.ResultChan <- Result{Text: "Please enter a valid number.", State: c.updateCarSaveYear}
		return
	}
	if _, err := c.updateDraftCarInDB(ctx, pl.UserID); err != nil {
		pl.ResultChan <- Result{Text: "Update failed, try again.", Error: err}
		return
	}
	res := Result{Text: "Car year has been successfully updated!"}
	car := c.draftCars[pl.UserID]
	res.InlineMarkup.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, car.ID))
	pl.ResultChan <- res
}

func (c *CarCommand) updateCarSavePlate(ctx context.Context, pl Payload) {
	c.setDraftCarPlate(pl.UserID, pl.Command)
	if _, err := c.updateDraftCarInDB(ctx, pl.UserID); err != nil {
		pl.ResultChan <- Result{Text: "Update failed, try again.", Error: err}
		return
	}
	res := Result{Text: "Car plate has been successfully updated!"}
	car := c.draftCars[pl.UserID]
	res.InlineMarkup.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, car.ID))
	pl.ResultChan <- res
}

func (c *CarCommand) updateCarSavePrice(ctx context.Context, pl Payload) {
	if c.setDraftCarPrice(pl.UserID, pl.Command) != nil {
		pl.ResultChan <- Result{Text: "Please enter a valid whole number.", State: c.updateCarSavePrice}
		return
	}
	if _, err := c.updateDraftCarInDB(ctx, pl.UserID); err != nil {
		pl.ResultChan <- Result{Text: "Update failed, try again.", Error: err}
		return
	}
	res := Result{Text: "Car price has been successfully updated!"}
	car := c.draftCars[pl.UserID]
	res.InlineMarkup.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, car.ID))
	pl.ResultChan <- res
}

func (c *CarCommand) deleteCarAsk(ctx context.Context, pl Payload, carID int64) {
	car, err := c.storage.GetCarFromDB(ctx, pl.UserID, carID)
	if err != nil {
		pl.ResultChan <- Result{Text: "Car not found.", Error: err}
		return
	}
	res := Result{Text: fmt.Sprintf("Are you sure you want to delete %s (%d)?", _es(car.Name), car.Year)}
	res.InlineMarkup.AddKeyboardButton("Yes, delete the car", commandf(c, cmdCarDelYes, carID))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("No", commandf(c, cmdCarGet, carID))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Nope, nevermind", commandf(c, cmdCarGet, carID))
	pl.ResultChan <- res
}

func (c *CarCommand) deleteCarConfirm(ctx context.Context, pl Payload, carID int64) {
	res := Result{}
	affected, err := c.storage.DeleteCarFromDB(ctx, pl.UserID, carID)
	if err != nil || affected != 1 {
		res.Text, res.Error = "Car not found.", err
	} else {
		res.Text = "Car has been successfully deleted!"
	}
	res.InlineMarkup.AddKeyboardButton("« Back to my cars", c.Prefix())
	pl.ResultChan <- res
}

func (c *CarCommand) formatFuelDetails(fuel st.FuelDetails) string {
	str := fmt.Sprintf("⛽ <b>Liters:</b> %.2fL (%s)\n", fuel.GetLiters(), fuel.Type)
	str += fmt.Sprintf("💲 <b>Paid:</b> %.2f€ (%.2fEur/L)\n", fuel.GetEuro(), fuel.GetEurPerLiter())
	str += fmt.Sprintf("📍 <b>Traveled:</b> %dKm (%.2fL/100Km)\n", fuel.KilometersR, fuel.GetLitersPerKilometer())
	str += fmt.Sprintf("🏭 <b>Total:</b> %dKm\n", fuel.Kilometers)
	str += fmt.Sprintf("📅 %s\n", fuel.GetTimestamp())
	return str
}

func (c *CarCommand) showFuelDetails(ctx context.Context, pl Payload, carID int64, offset int64) {
	res := Result{}
	fuel, err := c.storage.GetFuelFromDB(ctx, pl.UserID, carID, offset)
	if errors.Is(err, sql.ErrNoRows) {
		res.Text = "No fuel receipts found."
	} else if err != nil {
		res.Text, res.Error = "There is something wrong, please try again.", err
	} else {
		res.Text = c.formatFuelDetails(fuel)
		res.InlineMarkup.AddKeyboardPagination(offset, fuel.CountRows, commandf(c, cmdCarFuelGet, carID))
		res.InlineMarkup.AddKeyboardRow()
		res.InlineMarkup.AddKeyboardButton("Delete", commandf(c, cmdCarFuelDelAsk, carID, fuel.ID))
	}
	res.InlineMarkup.AddKeyboardButton("Add", commandf(c, cmdCarFuelAdd, carID))
	res.InlineMarkup.AddKeyboardRow()
	car, _ := c.storage.GetCarFromDB(ctx, pl.UserID, carID)
	res.InlineMarkup.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, carID))
	pl.ResultChan <- res
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
	c.draftFuel[userID].Milliliters = int64(math.Round(liters * 1000))
	return err
}

func (c *CarCommand) setDraftFuelKilometers(userID int64, input string) error {
	kilometers, err := strconv.Atoi(input)
	c.draftFuel[userID].Kilometers = int64(kilometers)
	return err
}

func (c *CarCommand) setDraftFuelEuros(userID int64, input string) error {
	euro, err := strconv.ParseFloat(input, 64)
	c.draftFuel[userID].Cents = int64(math.Round(euro * 100))
	return err
}

func (c *CarCommand) addFuelStart(ctx context.Context, pl Payload, carID int64) {
	c.newDraftFuel(pl.UserID, carID)
	res := Result{Text: "Please pick a receipt date.", State: c.addFuelTimestamp}
	res.InlineMarkup.AddKeyboardCalendar(time.Now().Year(), time.Now().Month())
	pl.ResultChan <- res
}

func (c *CarCommand) addFuelTimestamp(ctx context.Context, pl Payload) {
	res := Result{}
	if res.InlineMarkup.UpdateKeyboardCalendar(pl.Command) {
		pl.ResultChan <- res // new month is selected
		return
	}
	if c.setDraftFuelTimestamp(pl.UserID, pl.Command) != nil {
		pl.ResultChan <- Result{Text: "Please pick a date from the calendar.", State: c.addFuelTimestamp}
		return
	}
	res.Text = "Date: " + c.draftFuel[pl.UserID].GetTimestamp()
	res.InlineMarkup.AddKeyboardRow() // remove calendar keyboard
	pl.ResultChan <- res
	res = Result{Text: "What is the fuel type?", State: c.addFuelType}
	pl.ResultChan <- res
}

func (c *CarCommand) addFuelType(ctx context.Context, pl Payload) {
	c.setDraftFuelType(pl.UserID, pl.Command)
	pl.ResultChan <- Result{Text: "What is the fuel amount in Liters?", State: c.addFuelLiters}
}

func (c *CarCommand) addFuelLiters(ctx context.Context, pl Payload) {
	if err := c.setDraftFuelLiters(pl.UserID, pl.Command); err != nil {
		pl.ResultChan <- Result{Text: "Please enter a valid decimal number.", State: c.addFuelLiters}
	} else {
		pl.ResultChan <- Result{Text: "What is your total mileage now in Kilometers?", State: c.addFuelKilometers}
	}
}

func (c *CarCommand) addFuelKilometers(ctx context.Context, pl Payload) {
	if err := c.setDraftFuelKilometers(pl.UserID, pl.Command); err != nil {
		pl.ResultChan <- Result{Text: "Please enter a valid whole number.", State: c.addFuelKilometers}
	} else {
		pl.ResultChan <- Result{Text: "How much money did you spend in Euros?", State: c.addFuelEurosAndSave}
	}
}

func (c *CarCommand) addFuelEurosAndSave(ctx context.Context, pl Payload) {
	if err := c.setDraftFuelEuros(pl.UserID, pl.Command); err != nil {
		pl.ResultChan <- Result{Text: "Please enter a valid decimal number.", State: c.addFuelEurosAndSave}
		return
	}
	if _, err := c.insertDraftFuelIntoDB(ctx, pl.UserID); err != nil {
		pl.ResultChan <- Result{Text: "There is something wrong, please try again.", Error: err}
		return
	}
	c.showFuelDetails(ctx, pl, c.draftFuel[pl.UserID].CarID, 0)
}

func (c *CarCommand) deleteFuelAsk(ctx context.Context, pl Payload, carID int64, fuelID int64) {
	res := Result{Text: "Are you sure you want to delete the selected receipt?"}
	res.InlineMarkup.AddKeyboardButton("Yes, delete the receipt", commandf(c, cmdCarFuelDelYes, carID, fuelID))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("No", commandf(c, cmdCarFuelGet, carID))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Nope, nevermind", commandf(c, cmdCarFuelGet, carID))
	pl.ResultChan <- res
}

func (c *CarCommand) deleteFuelConfirm(ctx context.Context, pl Payload, carID int64, fuelID int64) {
	res := Result{}
	affected, err := c.storage.DeleteFuelFromDB(ctx, pl.UserID, fuelID)
	if err != nil || affected != 1 {
		res.Text, res.Error = "Receipt not found.", err
	} else {
		res.Text = "Receipt has been successfully deleted!"
	}
	res.InlineMarkup.AddKeyboardButton("« Back to my receipts", commandf(c, cmdCarFuelGet, carID))
	pl.ResultChan <- res
}

func (c *CarCommand) formatServiceDetails(service st.ServiceDetails) string {
	str := fmt.Sprintf("🛠️ %s\n", _es(service.Description))
	str += fmt.Sprintf("💲 <b>Paid:</b> %.2f€\n", service.GetEuro())
	str += fmt.Sprintf("📅 %s\n", service.GetTimestamp())
	return str
}

func (c *CarCommand) showServiceDetails(ctx context.Context, pl Payload, carID int64, offset int64) {
	res := Result{}
	service, err := c.storage.GetServiceFromDB(ctx, pl.UserID, carID, offset)
	if errors.Is(err, sql.ErrNoRows) {
		res.Text = "No service receipts found."
	} else if err != nil {
		res.Text, res.Error = "There is something wrong, please try again.", err
	} else {
		res.Text = c.formatServiceDetails(service)
		res.InlineMarkup.AddKeyboardPagination(offset, service.CountRows, commandf(c, cmdCarServiceGet, carID))
		res.InlineMarkup.AddKeyboardRow()
		res.InlineMarkup.AddKeyboardButton("Delete", commandf(c, cmdCarServiceDelAsk, carID, service.ID))
	}
	res.InlineMarkup.AddKeyboardButton("Add", commandf(c, cmdCarServiceAdd, carID))
	res.InlineMarkup.AddKeyboardRow()
	car, _ := c.storage.GetCarFromDB(ctx, pl.UserID, carID)
	res.InlineMarkup.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, carID))
	pl.ResultChan <- res
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
	c.draftService[userID].Cents = int64(math.Round(euro * 100))
	return err
}

func (c *CarCommand) addServiceStart(ctx context.Context, pl Payload, carID int64) {
	c.newDraftService(pl.UserID, carID)
	res := Result{Text: "Please pick a receipt date.", State: c.addServiceTimestamp}
	res.InlineMarkup.AddKeyboardCalendar(time.Now().Year(), time.Now().Month())
	pl.ResultChan <- res
}

func (c *CarCommand) addServiceTimestamp(ctx context.Context, pl Payload) {
	res := Result{}
	if res.InlineMarkup.UpdateKeyboardCalendar(pl.Command) {
		pl.ResultChan <- res // new month is selected
		return
	}
	if c.setDraftServiceTimestamp(pl.UserID, pl.Command) != nil {
		pl.ResultChan <- Result{Text: "Please pick a date from the calendar.", State: c.addServiceTimestamp}
		return
	}
	res.Text = "Date: " + c.draftService[pl.UserID].GetTimestamp()
	res.InlineMarkup.AddKeyboardRow() // remove calendar keyboard
	pl.ResultChan <- res
	res = Result{Text: "Provide service description.", State: c.addServiceDescription}
	pl.ResultChan <- res
}

func (c *CarCommand) addServiceDescription(ctx context.Context, pl Payload) {
	c.setDraftServiceDescription(pl.UserID, pl.Command)
	pl.ResultChan <- Result{Text: "How much money did you spend in Euros?", State: c.addServiceEurosAndSave}
}

func (c *CarCommand) addServiceEurosAndSave(ctx context.Context, pl Payload) {
	if err := c.setDraftServiceEuros(pl.UserID, pl.Command); err != nil {
		pl.ResultChan <- Result{Text: "Please enter a valid decimal number.", State: c.addServiceEurosAndSave}
		return
	}
	if _, err := c.insertDraftServiceIntoDB(ctx, pl.UserID); err != nil {
		pl.ResultChan <- Result{Text: "There is something wrong, please try again.", Error: err}
		return
	}
	c.showServiceDetails(ctx, pl, c.draftService[pl.UserID].CarID, 0)
}

func (c *CarCommand) deleteServiceAsk(ctx context.Context, pl Payload, carID int64, serviceID int64) {
	res := Result{Text: "Are you sure you want to delete the selected receipt?"}
	res.InlineMarkup.AddKeyboardButton("Yes, delete the receipt", commandf(c, cmdCarServiceDelYes, carID, serviceID))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("No", commandf(c, cmdCarServiceGet, carID))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Nope, nevermind", commandf(c, cmdCarServiceGet, carID))
	pl.ResultChan <- res
}

func (c *CarCommand) deleteServiceConfirm(ctx context.Context, pl Payload, carID int64, serviceID int64) {
	res := Result{}
	affected, err := c.storage.DeleteServiceFromDB(ctx, pl.UserID, serviceID)
	if err != nil || affected != 1 {
		res.Text, res.Error = "Receipt not found.", err
	} else {
		res.Text = "Receipt has been successfully deleted!"
	}
	res.InlineMarkup.AddKeyboardButton("« Back to my receipts", commandf(c, cmdCarServiceGet, carID))
	pl.ResultChan <- res
}

func (c *CarCommand) formatLeaseDetails(lease st.LeaseDetails) string {
	str := fmt.Sprintf("💲 <b>Paid:</b> %.2f€ (%.2f€ RT)\n", lease.GetEuro(), lease.GetEuroRT())
	if lease.Description.Valid {
		str += fmt.Sprintf("🛠️ %s\n", _es(lease.Description.String))
	}
	str += fmt.Sprintf("📅 %s\n", lease.GetTimestamp())
	return str
}

func (c *CarCommand) showLeaseDetails(ctx context.Context, pl Payload, carID int64, offset int64) {
	res := Result{}
	lease, err := c.storage.GetLeaseFromDB(ctx, pl.UserID, carID, offset)
	if errors.Is(err, sql.ErrNoRows) {
		res.Text = "No lease receipts found."
	} else if err != nil {
		res.Text, res.Error = "There is something wrong, please try again.", err
	} else {
		res.Text = c.formatLeaseDetails(lease)
		res.InlineMarkup.AddKeyboardPagination(offset, lease.CountRows, commandf(c, cmdCarLeaseGet, carID))
		res.InlineMarkup.AddKeyboardRow()
		res.InlineMarkup.AddKeyboardButton("Delete", commandf(c, cmdCarLeaseDelAsk, carID, lease.ID))
	}
	res.InlineMarkup.AddKeyboardButton("Add", commandf(c, cmdCarLeaseAdd, carID))
	res.InlineMarkup.AddKeyboardRow()
	car, _ := c.storage.GetCarFromDB(ctx, pl.UserID, carID)
	res.InlineMarkup.AddKeyboardButton(fmt.Sprintf("« Back to %s (%d)", car.Name, car.Year), commandf(c, cmdCarGet, carID))
	pl.ResultChan <- res
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
	if input == "/skip" {
		c.draftLease[userID].Description.Valid = false
	} else {
		c.draftLease[userID].Description.Valid = true
		c.draftLease[userID].Description.String = input
	}
}

func (c *CarCommand) setDraftLeaseEuros(userID int64, input string) error {
	euro, err := strconv.ParseFloat(input, 64)
	c.draftLease[userID].Cents = int64(math.Round(euro * 100))
	return err
}

func (c *CarCommand) addLeaseStart(ctx context.Context, pl Payload, carID int64) {
	c.newDraftLease(pl.UserID, carID)
	res := Result{Text: "Please pick a receipt date.", State: c.addLeaseTimestamp}
	res.InlineMarkup.AddKeyboardCalendar(time.Now().Year(), time.Now().Month())
	pl.ResultChan <- res
}

func (c *CarCommand) addLeaseTimestamp(ctx context.Context, pl Payload) {
	res := Result{}
	if res.InlineMarkup.UpdateKeyboardCalendar(pl.Command) {
		pl.ResultChan <- res // new month is selected
		return
	}
	if c.setDraftLeaseTimestamp(pl.UserID, pl.Command) != nil {
		pl.ResultChan <- Result{Text: "Please pick a date from the calendar.", State: c.addLeaseTimestamp}
		return
	}
	res.Text = "Date: " + c.draftLease[pl.UserID].GetTimestamp()
	res.InlineMarkup.AddKeyboardRow() // remove calendar keyboard
	pl.ResultChan <- res
	res = Result{Text: "Provide lease description. /skip", State: c.addLeaseDescription}
	pl.ResultChan <- res
}

func (c *CarCommand) addLeaseDescription(ctx context.Context, pl Payload) {
	c.setDraftLeaseDescription(pl.UserID, pl.Command)
	pl.ResultChan <- Result{Text: "How much money did you spend in Euros?", State: c.addLeaseEurosAndSave}
}

func (c *CarCommand) addLeaseEurosAndSave(ctx context.Context, pl Payload) {
	if err := c.setDraftLeaseEuros(pl.UserID, pl.Command); err != nil {
		pl.ResultChan <- Result{Text: "Please enter a valid decimal number.", State: c.addLeaseEurosAndSave}
		return
	}
	if _, err := c.insertDraftLeaseIntoDB(ctx, pl.UserID); err != nil {
		pl.ResultChan <- Result{Text: "There is something wrong, please try again.", Error: err}
		return
	}
	c.showLeaseDetails(ctx, pl, c.draftLease[pl.UserID].CarID, 0)
}

func (c *CarCommand) deleteLeaseAsk(ctx context.Context, pl Payload, carID int64, leaseID int64) {
	res := Result{Text: "Are you sure you want to delete the selected receipt?"}
	res.InlineMarkup.AddKeyboardButton("Yes, delete the receipt", commandf(c, cmdCarLeaseDelYes, carID, leaseID))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("No", commandf(c, cmdCarLeaseGet, carID))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Nope, nevermind", commandf(c, cmdCarLeaseGet, carID))
	pl.ResultChan <- res
}

func (c *CarCommand) deleteLeaseConfirm(ctx context.Context, pl Payload, carID int64, leaseID int64) {
	res := Result{}
	affected, err := c.storage.DeleteLeaseFromDB(ctx, pl.UserID, leaseID)
	if err != nil || affected != 1 {
		res.Text, res.Error = "Receipt not found.", err
	} else {
		res.Text = "Receipt has been successfully deleted!"
	}
	res.InlineMarkup.AddKeyboardButton("« Back to my receipts", commandf(c, cmdCarLeaseGet, carID))
	pl.ResultChan <- res
}
