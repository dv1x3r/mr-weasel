package tgmanager

import (
	"testing"
)

func TestResultKeyboard(t *testing.T) {
	res := Result{}
	if res.Keyboard != nil {
		t.Fatal("Keyboard is not nil on init")
	}
	for i := 0; i < 5; i++ {
		res.AddKeyboardRow()
		if len(res.Keyboard.InlineKeyboard) != i {
			t.Errorf("actual rows: [%d], expected rows [%d]\n", len(res.Keyboard.InlineKeyboard), i)
		}
	}
}
