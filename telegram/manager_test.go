package telegram

import (
	"context"
	"mr-weasel/commands"
	"testing"
)

type TestCommand struct {
}

func (TestCommand) Prefix() string {
	return "test"
}

func (TestCommand) Description() string {
	return "Test Command"
}

func (tc *TestCommand) Execute(ctx context.Context, pl commands.Payload) (commands.Result, error) {
	if pl.Command == "state" {
		return commands.Result{State: tc.ExecuteState}, nil
	}
	return commands.Result{}, nil
}

func (tc *TestCommand) ExecuteState(ctx context.Context, pl commands.Payload) (commands.Result, error) {
	return commands.Result{}, nil
}

func TestGetHandlerFunc(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		manager := NewManager(nil)
		fn, ok := manager.getHandlerFunc(0, "/test")
		if fn != nil || ok {
			t.Fatalf("expected [nil, false], actual [%p %t]\n", fn, ok)
		}
	})

	t.Run("Found", func(t *testing.T) {
		manager := NewManager(nil)
		manager.AddCommands(new(TestCommand))
		fn, ok := manager.getHandlerFunc(0, "/test")
		if fn == nil || !ok {
			t.Fatalf("expected [nil, false], actual [%p %t]\n", fn, ok)
		}
	})
}

func TestCommandKeyboardToInlineMarkup(t *testing.T) {
	keyboard := [][]commands.Button{
		{
			{Label: "ABC", Data: "DEF"},
		},
		{
			{Label: "GHI", Data: "JKL"},
			{Label: "MNO", Data: "PQR"},
		},
	}
	markup := NewManager(nil).commandKeyboardToInlineMarkup(keyboard)
	for r := 0; r < len(keyboard); r++ {
		for b := 0; b < len(keyboard[r]); b++ {
			testLabel, testData := keyboard[r][b].Label, keyboard[r][b].Data
			markLabel, markData := markup.InlineKeyboard[r][b].Text, markup.InlineKeyboard[r][b].CallbackData
			if testLabel != markLabel || testData != markData {
				t.Errorf("expected [%s, %s] button, actual [%s, %s]\n", testLabel, testData, markLabel, markData)
			}
		}
	}
}
