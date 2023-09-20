package telegram

import (
	"context"
	"testing"

	"mr-weasel/commands"
)

type TestCommand struct {
}

func (TestCommand) Prefix() string {
	return "/test"
}

func (TestCommand) Description() string {
	return "Test Command"
}

func (tc *TestCommand) Execute(ctx context.Context, pl commands.Payload) {
	if pl.Command == "state" {
		pl.ResultChan <- commands.Result{State: tc.ExecuteState}
	}
	pl.ResultChan <- commands.Result{}
}

func (tc *TestCommand) ExecuteState(ctx context.Context, pl commands.Payload) {
	pl.ResultChan <- commands.Result{}
}

func TestGetHandlerFunc(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		manager := NewManager(nil)
		fn, ok := manager.getExecuteFunc(0, "/test")
		if fn != nil || ok {
			t.Fatalf("expected [nil, false], actual [%p %t]\n", fn, ok)
		}
	})

	t.Run("Found", func(t *testing.T) {
		manager := NewManager(nil)
		manager.AddCommands(new(TestCommand))
		fn, ok := manager.getExecuteFunc(0, "/test")
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
