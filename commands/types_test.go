package commands

import "testing"

func TestKeyboard(t *testing.T) {
	t.Run("AddRow", func(t *testing.T) {
		res := Result{}
		res.AddKeyboardRow()
		if len(res.Keyboard.InlineKeyboard) != 1 {
			t.Fatalf("expected [1] row, actual [%d]\n", len(res.Keyboard.InlineKeyboard))
		}
	})
	t.Run("AddButtonNoRow", func(t *testing.T) {
		res := Result{}
		res.AddKeyboardButton("abc", "def")
		button := res.Keyboard.InlineKeyboard[0][0]
		if button.Text != "abc" || button.CallbackData != "def" {
			t.Fatalf("expected [abc, def] button, actual [%s, %s]\n", button.Text, button.CallbackData)
		}
	})
	t.Run("AddButtonExistingRow", func(t *testing.T) {
		res := Result{}
		res.AddKeyboardRow()
		res.AddKeyboardButton("abc", "def")
		button := res.Keyboard.InlineKeyboard[0][0]
		if button.Text != "abc" || button.CallbackData != "def" {
			t.Fatalf("expected [abc, def] button, actual [%s, %s]\n", button.Text, button.CallbackData)
		}
	})
}
