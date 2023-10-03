package commands

import (
	"context"
	"mr-weasel/tgclient"
)

const CmdCancel = "/cancel"

type ExecuteFunc = func(context.Context, Payload)

type Handler interface {
	Prefix() string
	Description() string
	Execute(context.Context, Payload)
}

type Payload struct {
	UserID     int64
	Command    string
	FileURL    string
	ResultChan chan Result
}

type Result struct {
	Text         string
	State        ExecuteFunc
	InlineMarkup tgclient.InlineKeyboardMarkup
	ReplyMarkup  tgclient.ReplyKeyboardMarkup
	RemoveMarkup tgclient.ReplyKeyboardRemove
	Audio        map[string]string
	ClearState   bool
	Error        error
}
