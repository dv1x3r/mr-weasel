package commands

import (
	"context"
	"fmt"

	st "mr-weasel/storage"
)

type HolidayCommand struct {
	storage *st.HolidayStorage
}

func NewHolidayCommand(storage *st.HolidayStorage) *HolidayCommand {
	return &HolidayCommand{storage: storage}
}

func (HolidayCommand) Prefix() string {
	return "/holiday"
}

func (HolidayCommand) Description() string {
	return "manage day off's"
}

func (c *HolidayCommand) Execute(ctx context.Context, pl Payload) (Result, error) {
	return c.showHolodayDaysByYear(ctx, pl.UserID)
}

func (c *HolidayCommand) showHolodayDaysByYear(ctx context.Context, userID int64) (Result, error) {
	holidays, err := c.storage.SelectHolidayDaysByYearFromDB(ctx, userID)
	if err != nil {
		return Result{Text: "There is something wrong, please try again."}, err
	}

	res := Result{}
	if len(holidays) == 0 {
		res.Text = "Holidays not found, add one?"
	} else {
		res.Text = "Holiday days by year:"
		for _, v := range holidays {
			res.Text += fmt.Sprintf("\n<b>%d</b> - %d days", v.Year, v.Days)
		}
	}

	res.AddKeyboardButton("Show details", "-")

	return res, nil
}
