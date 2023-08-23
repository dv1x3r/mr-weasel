package commands

import (
	"context"

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
	return Result{}, nil
}
