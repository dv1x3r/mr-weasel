package tgclient

import "testing"

func TestInlineKeyboardMarkup(t *testing.T) {
	t.Run("AddRow", func(t *testing.T) {
		markup := InlineKeyboardMarkup{}
		markup.AddRow()
		if len(markup.InlineKeyboard) != 1 {
			t.Fatalf("expected [1] row, actual [%d]\n", len(markup.InlineKeyboard))
		}
	})
	t.Run("AddButtonNoRow", func(t *testing.T) {
		markup := InlineKeyboardMarkup{}
		markup.AddButton("abc", "def")
		button := markup.InlineKeyboard[0][0]
		if button.Text != "abc" || button.CallbackData != "def" {
			t.Fatalf("expected [abc, def] button, actual [%s, %s]\n", button.Text, button.CallbackData)
		}
	})
	t.Run("AddButtonExistingRow", func(t *testing.T) {
		markup := InlineKeyboardMarkup{}
		markup.AddRow()
		markup.AddButton("abc", "def")
		button := markup.InlineKeyboard[0][0]
		if button.Text != "abc" || button.CallbackData != "def" {
			t.Fatalf("expected [abc, def] button, actual [%s, %s]\n", button.Text, button.CallbackData)
		}
	})
}
