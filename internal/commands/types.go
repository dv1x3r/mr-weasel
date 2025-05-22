package commands

import (
	"context"
	"html"
	"mr-weasel/internal/lib/telegram"
)

var _es = html.EscapeString

const CmdCancel = "/cancel"

type ExecuteFunc = func(context.Context, Payload)

type Handler interface {
	Prefix() string
	Description() string
	Execute(context.Context, Payload)
}

type Payload struct {
	UserID     int64
	UserName   string
	IsPrivate  bool
	Command    string
	FileURL    string
	ResultChan chan Result
}

type Result struct {
	Text         string
	State        ExecuteFunc
	InlineMarkup telegram.InlineKeyboardMarkup
	ReplyMarkup  telegram.ReplyKeyboardMarkup
	RemoveMarkup telegram.ReplyKeyboardRemove
	Audio        map[string]string
	ClearState   bool
	Error        error
}
