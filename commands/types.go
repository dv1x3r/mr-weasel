package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Handler interface {
	Prefix() string
	Description() string
	Execute(context.Context, Payload) (Result, error)
}

type HandlerFunc = func(context.Context, Payload) (Result, error)

type Payload struct {
	UserID  int64
	Command string
}

type Result struct {
	Text     string
	State    HandlerFunc
	Keyboard [][]Button
}

type Button struct {
	Label string
	Data  string
}

func (r *Result) AddKeyboardRow() {
	if r.Keyboard == nil {
		r.Keyboard = make([][]Button, 1)
	} else {
		r.Keyboard = append(r.Keyboard, make([]Button, 0))
	}
}

func (r *Result) AddKeyboardButton(label string, data string) {
	if r.Keyboard == nil {
		r.AddKeyboardRow()
	}
	i := len(r.Keyboard) - 1
	r.Keyboard[i] = append(r.Keyboard[i], Button{Label: label, Data: data})
}

func (r *Result) AddKeyboardCalendar(year int, month time.Month) {
	ISOWeekday := func(t time.Time) int {
		if t.Weekday() == time.Sunday {
			return 6
		} else {
			return int(t.Weekday()) - 1
		}
	}

	dt := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	yearMonthStr := fmt.Sprintf("%s %d", month.String(), year)

	r.AddKeyboardButton(yearMonthStr, "-")
	r.AddKeyboardRow()
	for _, v := range []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"} {
		r.AddKeyboardButton(v, "-")
	}

	// in case month starts not on Monday, add empty buttons
	r.AddKeyboardRow()
	for i := 0; i < ISOWeekday(dt); i++ {
		r.AddKeyboardButton(" ", "-")
	}

	for day := 1; day <= daysInMonth; day++ {
		dt = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		r.AddKeyboardButton(fmt.Sprint(day), fmt.Sprint(dt.Unix()))
		if ISOWeekday(dt) == 6 {
			r.AddKeyboardRow()
		}
	}

	// in case month ends not on Sunday, add empty buttons
	for n := ISOWeekday(dt); n < 6; n++ {
		r.AddKeyboardButton(" ", "-")
	}

	dtPrev := time.Date(dt.Year(), dt.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	dtNext := time.Date(dt.Year(), dt.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	r.AddKeyboardRow()
	r.AddKeyboardButton("«", fmt.Sprintf("%d %d", dtPrev.Year(), dtPrev.Month()))
	r.AddKeyboardButton("Pick Today", fmt.Sprint(time.Now().Unix()))
	r.AddKeyboardButton("»", fmt.Sprintf("%d %d", dtNext.Year(), dtNext.Month()))
}

func splitCommand(input string, prefix string) []string {
	input, _ = strings.CutPrefix(input, prefix)
	input = strings.TrimSpace(input)
	if input == "" {
		return []string{}
	}
	return strings.Split(input, " ")
}

func commandf(h Handler, args ...any) string {
	cmd := h.Prefix()
	for _, arg := range args {
		cmd = fmt.Sprintf("%s %v", cmd, arg)
	}
	return cmd
}

func safeGet(args []string, n int) string {
	if n <= len(args)-1 {
		return args[n]
	}
	return ""
}

func safeGetInt(args []string, n int) int {
	i, _ := strconv.Atoi(safeGet(args, n))
	return i
}

func safeGetInt64(args []string, n int) int64 {
	return int64(safeGetInt(args, n))
}
