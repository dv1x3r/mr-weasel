package commands

import "testing"

func TestKeyboard(t *testing.T) {
	t.Run("AddRow", func(t *testing.T) {
		res := Result{}
		res.AddKeyboardRow()
		if len(res.Keyboard) != 1 {
			t.Fatalf("expected [1] row, actual [%d]\n", len(res.Keyboard))
		}
	})
	t.Run("AddButtonNoRow", func(t *testing.T) {
		res := Result{}
		res.AddKeyboardButton("abc", "def")
		button := res.Keyboard[0][0]
		if button.Label != "abc" || button.Data != "def" {
			t.Fatalf("expected [abc, def] button, actual [%s, %s]\n", button.Label, button.Data)
		}
	})
	t.Run("AddButtonExistingRow", func(t *testing.T) {
		res := Result{}
		res.AddKeyboardRow()
		res.AddKeyboardButton("abc", "def")
		button := res.Keyboard[0][0]
		if button.Label != "abc" || button.Data != "def" {
			t.Fatalf("expected [abc, def] button, actual [%s, %s]\n", button.Label, button.Data)
		}
	})
}
