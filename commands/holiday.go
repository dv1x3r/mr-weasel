package commands

import (
	"context"
	"fmt"
	"strconv"
	"time"

	st "mr-weasel/storage"
)

type HolidayCommand struct {
	storage       *st.HolidayStorage
	draftHolidays map[int64]*st.HolidayBase
}

func NewHolidayCommand(storage *st.HolidayStorage) *HolidayCommand {
	return &HolidayCommand{
		storage:       storage,
		draftHolidays: make(map[int64]*st.HolidayBase),
	}
}

func (HolidayCommand) Prefix() string {
	return "/holiday"
}

func (HolidayCommand) Description() string {
	return "day off's storage"
}

const (
	cmdHolidayAdd    = "add"
	cmdHolidayGet    = "get"
	cmdHolidayDelAsk = "del"
	cmdHolidayDelYes = "del_yes"
)

func (c *HolidayCommand) Execute(ctx context.Context, pl Payload) (Result, error) {
	args := splitCommand(pl.Command, c.Prefix())
	switch safeGet(args, 0) {
	case cmdHolidayAdd:
		return c.addHolidayStart(ctx, pl.UserID)
	// case cmdHolidayGet:
	// 	return c.showFuelDetails(ctx, pl.UserID, safeGetInt64(args, 1), safeGetInt64(args, 2))
	// case cmdHolidayDelAsk:
	// 	return c.deleteFuelAsk(ctx, pl.UserID, safeGetInt64(args, 1), safeGetInt64(args, 2))
	// case cmdHolidayDelYes:
	// 	return c.deleteFuelConfirm(ctx, pl.UserID, safeGetInt64(args, 1), safeGetInt64(args, 2))
	default:
		return c.showHolodayDaysByYear(ctx, pl.UserID)
	}
}

func (c *HolidayCommand) showHolidayDetails(ctx context.Context, userID int64, holidayID int64) (Result, error) {
	return Result{}, nil
}

func (c *HolidayCommand) showHolodayDaysByYear(ctx context.Context, userID int64) (Result, error) {
	holidays, err := c.storage.SelectHolidayDaysByYearFromDB(ctx, userID)
	if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}

	res := Result{}

	if len(holidays) == 0 {
		res.Text = "Holidays not found, add one?"
		res.AddKeyboardButton("Add", commandf(c, cmdHolidayAdd))
	} else {
		res.Text = "Holiday days by year:"
		res.AddKeyboardButton("Manage", commandf(c, cmdHolidayGet))
		for _, v := range holidays {
			res.Text += fmt.Sprintf("\n<b>%d</b> - %d days", v.Year, v.Days)
		}
	}

	return res, nil
}

func (c *HolidayCommand) insertDraftHolidayIntoDB(ctx context.Context, userID int64) (int64, error) {
	return c.storage.InsertHolidayIntoDB(ctx, *c.draftHolidays[userID])
}

func (c *HolidayCommand) newDraftHoliday(userID int64) {
	c.draftHolidays[userID] = &st.HolidayBase{UserID: userID}
}

func (c *HolidayCommand) setDraftHolidayStartDate(userID int64, input string) error {
	timestamp, err := strconv.Atoi(input)
	c.draftHolidays[userID].Start = int64(timestamp)
	return err
}

func (c *HolidayCommand) setDraftHolidayEndDate(userID int64, input string) error {
	timestamp, err := strconv.Atoi(input)
	c.draftHolidays[userID].End = int64(timestamp)
	return err
}

func (c *HolidayCommand) setDraftHolidayDays(userID int64, input string) error {
	days, err := strconv.Atoi(input)
	c.draftHolidays[userID].Days = int64(days)
	return err
}

func (c *HolidayCommand) addHolidayStart(ctx context.Context, userID int64) (Result, error) {
	c.newDraftHoliday(userID)
	res := Result{Text: "Please pick vacation start date.", State: c.addHolidayStartDate}
	res.AddKeyboardCalendar(time.Now().Year(), time.Now().Month())
	return res, nil
}

func (c *HolidayCommand) addHolidayStartDate(ctx context.Context, pl Payload) (Result, error) {
	res := Result{}
	if res.UpdateKeyboardCalendar(pl.Command) {
		return res, nil
	} else if c.setDraftHolidayStartDate(pl.UserID, pl.Command) != nil {
		return res, nil
	}

	res.Text, res.State = "Please pick vacation end date.", c.addHolidayEndDate
	res.AddKeyboardCalendar(time.Now().Year(), time.Now().Month())
	return res, nil
}

func (c *HolidayCommand) addHolidayEndDate(ctx context.Context, pl Payload) (Result, error) {
	res := Result{}
	if res.UpdateKeyboardCalendar(pl.Command) {
		return res, nil
	} else if c.setDraftHolidayEndDate(pl.UserID, pl.Command) != nil {
		return res, nil
	}

	res.Text, res.State = "Enter number of working days.", c.addHolidayDaysAndSave
	res.AddKeyboardRow() // remove calendar keyboard
	return res, nil
}

func (c *HolidayCommand) addHolidayDaysAndSave(ctx context.Context, pl Payload) (Result, error) {
	if c.setDraftHolidayDays(pl.UserID, pl.Command) != nil {
		return Result{Text: "Please enter a valid whole number.", State: c.addHolidayDaysAndSave}, nil
	}

	holidayID, err := c.insertDraftHolidayIntoDB(ctx, pl.UserID)
	if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}

	return c.showHolidayDetails(ctx, pl.UserID, holidayID)
}
