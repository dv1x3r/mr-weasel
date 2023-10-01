package tgclient

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (kb *InlineKeyboardMarkup) AddKeyboardRow() {
	if kb.InlineKeyboard == nil {
		kb.InlineKeyboard = make([][]InlineKeyboardButton, 1)
	} else {
		kb.InlineKeyboard = append(kb.InlineKeyboard, make([]InlineKeyboardButton, 0))
	}
}

func (kb *InlineKeyboardMarkup) AddKeyboardButton(text string, callbackData string) {
	if kb.InlineKeyboard == nil {
		kb.AddKeyboardRow()
	}
	i := len(kb.InlineKeyboard) - 1
	kb.InlineKeyboard[i] = append(kb.InlineKeyboard[i], InlineKeyboardButton{Text: text, CallbackData: callbackData})
}

func (kb *InlineKeyboardMarkup) AddKeyboardCalendar(year int, month time.Month) {
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

	kb.AddKeyboardButton(yearMonthStr, "-")
	kb.AddKeyboardRow()
	for _, v := range []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"} {
		kb.AddKeyboardButton(v, "-")
	}

	// in case month starts not on Monday, add empty buttons
	kb.AddKeyboardRow()
	for i := 0; i < ISOWeekday(dt); i++ {
		kb.AddKeyboardButton(" ", "-")
	}

	for day := 1; day <= daysInMonth; day++ {
		dt = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		kb.AddKeyboardButton(fmt.Sprint(day), fmt.Sprint(dt.Unix()))
		if ISOWeekday(dt) == 6 {
			kb.AddKeyboardRow()
		}
	}

	// in case month ends not on Sunday, add empty buttons
	for n := ISOWeekday(dt); n < 6; n++ {
		kb.AddKeyboardButton(" ", "-")
	}

	dtPrev := time.Date(dt.Year(), dt.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	dtNext := time.Date(dt.Year(), dt.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	kb.AddKeyboardRow()
	kb.AddKeyboardButton("«", fmt.Sprintf("%d %d", dtPrev.Year(), dtPrev.Month()))
	// kb.AddKeyboardButton("Pick Today", fmt.Sprint(time.Now().Unix()))
	kb.AddKeyboardButton("»", fmt.Sprintf("%d %d", dtNext.Year(), dtNext.Month()))
}

func (kb *InlineKeyboardMarkup) UpdateKeyboardCalendar(input string) bool {
	if s := strings.Split(input, " "); len(s) == 2 {
		year, _ := strconv.Atoi(s[0])
		month, _ := strconv.Atoi(s[1])
		kb.AddKeyboardCalendar(year, time.Month(month))
		return true
	}
	return false
}

func (kb *InlineKeyboardMarkup) AddKeyboardPagination(offset int64, countRows int64, command string) {
	if offset >= 5 {
		kb.AddKeyboardButton("«5", fmt.Sprintf("%s %d", command, offset-5))
	} else if offset == 0 {
		kb.AddKeyboardButton(" ", "-")
	} else {
		kb.AddKeyboardButton(fmt.Sprintf("«%d", offset), fmt.Sprintf("%s %d", command, 0))
	}
	if offset >= 1 {
		kb.AddKeyboardButton("«1", fmt.Sprintf("%s %d", command, offset-1))
	} else {
		kb.AddKeyboardButton(" ", "-")
	}
	kb.AddKeyboardButton(fmt.Sprintf("%d/%d", offset+1, countRows), "-")
	if offset+1 < countRows {
		kb.AddKeyboardButton("1»", fmt.Sprintf("%s %d", command, offset+1))
	} else {
		kb.AddKeyboardButton(" ", "-")
	}
	if offset+1 < countRows-4 {
		kb.AddKeyboardButton("5»", fmt.Sprintf("%s %d", command, offset+5))
	} else if offset == countRows-1 {
		kb.AddKeyboardButton(" ", "-")
	} else {
		kb.AddKeyboardButton(fmt.Sprintf("%d»", countRows-1-offset), fmt.Sprintf("%s %d", command, countRows-1))
	}
}
