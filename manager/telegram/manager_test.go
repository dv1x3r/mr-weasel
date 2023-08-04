package tgmanager

import (
	"testing"
)

func TestResultKeyboard(t *testing.T) {
	res := Result{}
	if res.Keyboard != nil {
		t.Fatal("Keyboard is not nil on Result init")
	}
	t.Run("NewRows", func(t *testing.T) {
		res := Result{}
		res.AddKeyboardRow()
		if len(res.Keyboard.InlineKeyboard) != 1 {
			t.Errorf("actual rows: [%d], expected rows [%d]\n", len(res.Keyboard.InlineKeyboard), 1)
		}
		res.AddKeyboardRow()
		if len(res.Keyboard.InlineKeyboard) != 2 {
			t.Errorf("actual rows: [%d], expected rows [%d]\n", len(res.Keyboard.InlineKeyboard), 2)
		}
	})
}
