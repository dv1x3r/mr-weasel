package commands

import (
	"context"
	"strings"
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

func splitCommand(input string, prefix string) []string {
	input, _ = strings.CutPrefix(input, prefix)
	input = strings.TrimSpace(input)
	if input == "" {
		return []string{}
	}
	return strings.Split(input, " ")
}

func safeGet(args []string, n int) string {
	if n <= len(args)-1 {
		return args[n]
	}
	return ""
}
