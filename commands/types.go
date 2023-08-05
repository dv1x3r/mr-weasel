package commands

import (
	"context"
	"mr-weasel/tgclient"
)

type Payload struct {
	UserID  int64
	Command string
}

type Result struct {
	Text     string
	State    HandlerFunc
	Keyboard *tgclient.InlineKeyboardMarkup
}

func (r *Result) AddKeyboardRow() {
	if r.Keyboard == nil {
		r.Keyboard = &tgclient.InlineKeyboardMarkup{}
	}
	r.Keyboard.InlineKeyboard = append(r.Keyboard.InlineKeyboard, []tgclient.InlineKeyboardButton{})
}

func (r *Result) AddKeyboardButton(label string, data string) {
	if r.Keyboard == nil {
		r.AddKeyboardRow()
	}
	row := &r.Keyboard.InlineKeyboard[len(r.Keyboard.InlineKeyboard)-1]
	*row = append(*row, tgclient.InlineKeyboardButton{Text: label, CallbackData: data})
}

type HandlerFunc = func(context.Context, Payload) (Result, error)

type Handler interface {
	Prefix() string
	Description() string
	Execute(context.Context, Payload) (Result, error)
}
