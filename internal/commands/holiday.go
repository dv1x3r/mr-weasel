package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	st "mr-weasel/internal/storage"
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
	return "manage holiday days"
}

const (
	cmdHolidayAdd    = "add"
	cmdHolidayGet    = "get"
	cmdHolidayDelAsk = "del"
	cmdHolidayDelYes = "del_yes"
)

func (c *HolidayCommand) Execute(ctx context.Context, pl Payload) {
	args := splitCommand(pl.Command, c.Prefix())
	switch safeGet(args, 0) {
	case cmdHolidayAdd:
		c.addHolidayStart(ctx, pl)
	case cmdHolidayGet:
		c.showHolidayDetails(ctx, pl, safeGetInt64(args, 1))
	case cmdHolidayDelAsk:
		c.deleteHolidayAsk(ctx, pl, safeGetInt64(args, 1))
	case cmdHolidayDelYes:
		c.deleteHolidayConfirm(ctx, pl, safeGetInt64(args, 1))
	default:
		c.showHolodayDaysByYear(ctx, pl)
	}
}

func (c *HolidayCommand) formatHolidayDetails(holiday st.HolidayDetails) string {
	html := fmt.Sprintf("📅 <b>Start:</b> %s\n", holiday.GetStartTimestamp())
	html += fmt.Sprintf("📅 <b>End:</b> %s\n", holiday.GetEndTimestamp())
	html += fmt.Sprintf("🌴 <b>Working days:</b> %d\n", holiday.Days)
	return html
}

func (c *HolidayCommand) showHolidayDetails(ctx context.Context, pl Payload, offset int64) {
	res := Result{}
	holiday, err := c.storage.GetHolidayFromDB(ctx, pl.UserID, offset)
	if errors.Is(err, sql.ErrNoRows) {
		res.Text = "Holiday not found."
	} else if err != nil {
		res.Text, res.Error = "There is something wrong, please try again.", err
	} else {
		res.Text = c.formatHolidayDetails(holiday)
		res.InlineMarkup.AddKeyboardPagination(offset, holiday.CountRows, commandf(c, cmdHolidayGet))
		res.InlineMarkup.AddKeyboardRow()
		res.InlineMarkup.AddKeyboardButton("Delete", commandf(c, cmdHolidayDelAsk, holiday.ID))
	}
	res.InlineMarkup.AddKeyboardButton("Add", commandf(c, cmdHolidayAdd))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("« Back", commandf(c))
	pl.ResultChan <- res
}

func (c *HolidayCommand) showHolodayDaysByYear(ctx context.Context, pl Payload) {
	res := Result{}
	holidays, err := c.storage.SelectHolidayDaysByYearFromDB(ctx, pl.UserID)
	if err != nil {
		res.Text, res.Error = "There is something wrong, please try again.", err
	} else if len(holidays) == 0 {
		res.Text = "Holidays not found, add one?"
		res.InlineMarkup.AddKeyboardButton("Add", commandf(c, cmdHolidayAdd))
	} else {
		res.Text = "Holiday days by year:"
		res.InlineMarkup.AddKeyboardButton("Manage", commandf(c, cmdHolidayGet))
		for _, v := range holidays {
			res.Text += fmt.Sprintf("\n<b>%d</b> - %d offline days", v.Year, v.Days)
		}
	}
	pl.ResultChan <- res
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

func (c *HolidayCommand) addHolidayStart(ctx context.Context, pl Payload) {
	c.newDraftHoliday(pl.UserID)
	res := Result{Text: "Please pick holiday start date.", State: c.addHolidayStartDate}
	res.InlineMarkup.AddKeyboardCalendar(time.Now().Year(), time.Now().Month())
	pl.ResultChan <- res
}

func (c *HolidayCommand) addHolidayStartDate(ctx context.Context, pl Payload) {
	res := Result{}
	if res.InlineMarkup.UpdateKeyboardCalendar(pl.Command) {
		pl.ResultChan <- res // new month is selected
		return
	}
	if c.setDraftHolidayStartDate(pl.UserID, pl.Command) != nil {
		pl.ResultChan <- Result{Text: "Please pick a date from the calendar.", State: c.addHolidayStartDate}
		return
	}
	res.Text = "Start: " + c.draftHolidays[pl.UserID].GetStartTimestamp()
	res.InlineMarkup.AddKeyboardRow() // remove calendar keyboard
	pl.ResultChan <- res
	res = Result{Text: "Please pick holiday end date.", State: c.addHolidayEndDate}
	res.InlineMarkup.AddKeyboardCalendar(time.Now().Year(), time.Now().Month())
	pl.ResultChan <- res
}

func (c *HolidayCommand) addHolidayEndDate(ctx context.Context, pl Payload) {
	res := Result{}
	if res.InlineMarkup.UpdateKeyboardCalendar(pl.Command) {
		pl.ResultChan <- res
		return
	}
	if c.setDraftHolidayEndDate(pl.UserID, pl.Command) != nil {
		pl.ResultChan <- res
		return
	}
	res.Text = "End: " + c.draftHolidays[pl.UserID].GetEndTimestamp()
	res.InlineMarkup.AddKeyboardRow() // remove calendar keyboard
	pl.ResultChan <- res
	res = Result{Text: "Enter number of working days.", State: c.addHolidayDaysAndSave}
	pl.ResultChan <- res
}

func (c *HolidayCommand) addHolidayDaysAndSave(ctx context.Context, pl Payload) {
	if c.setDraftHolidayDays(pl.UserID, pl.Command) != nil {
		pl.ResultChan <- Result{Text: "Please enter a valid whole number.", State: c.addHolidayDaysAndSave}
		return
	}
	_, err := c.insertDraftHolidayIntoDB(ctx, pl.UserID)
	if err != nil {
		pl.ResultChan <- Result{Text: "There is something wrong, please try again.", Error: err}
		return
	}
	c.showHolidayDetails(ctx, pl, 0)
}

func (c *HolidayCommand) deleteHolidayAsk(ctx context.Context, pl Payload, holidayID int64) {
	res := Result{Text: "Are you sure you want to delete the selected holiday?"}
	res.InlineMarkup.AddKeyboardButton("Yes, delete the holiday", commandf(c, cmdHolidayDelYes, holidayID))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("No", commandf(c, cmdHolidayGet))
	res.InlineMarkup.AddKeyboardRow()
	res.InlineMarkup.AddKeyboardButton("Nope, nevermind", commandf(c, cmdHolidayGet))
	pl.ResultChan <- res
}

func (c *HolidayCommand) deleteHolidayConfirm(ctx context.Context, pl Payload, holidayID int64) {
	affected, err := c.storage.DeleteHolidayFromDB(ctx, pl.UserID, holidayID)
	if err != nil || affected != 1 {
		pl.ResultChan <- Result{Text: "Holiday not found.", Error: err}
		return
	}
	res := Result{Text: "Holiday has been successfully deleted!"}
	res.InlineMarkup.AddKeyboardButton("« Back to my holidays", c.Prefix())
	pl.ResultChan <- res
}
