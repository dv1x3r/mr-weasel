package tgmanager

import (
	"context"
	"log"
	"mr-weasel/client/telegram"
	"strings"
)

type Payload struct {
	User            *tgclient.User
	Chat            *tgclient.Chat
	CallbackMessage *tgclient.Message
	Command         string
	ClearState      bool
}

type Result struct {
	Text     string
	State    HandlerFunc
	Keyboard *tgclient.InlineKeyboardMarkup
}

func (res *Result) AddKeyboardButton(text string, data string) {
	if res.Keyboard == nil {
		res.Keyboard = &tgclient.InlineKeyboardMarkup{
			InlineKeyboard: make([][]tgclient.InlineKeyboardButton, 1),
		}
	}
	row := &res.Keyboard.InlineKeyboard[len(res.Keyboard.InlineKeyboard)-1]
	*row = append(*row, tgclient.InlineKeyboardButton{Text: text, CallbackData: data})
}

func (res *Result) AddKeyboardRow() {
	if res.Keyboard != nil && res.Keyboard.InlineKeyboard != nil {
		res.Keyboard.InlineKeyboard = append(res.Keyboard.InlineKeyboard,
			[]tgclient.InlineKeyboardButton{})
	}
}

type Handler interface {
	Prefix() string
	Description() string
	Execute(context.Context, Payload) (Result, error)
}

type HandlerFunc = func(context.Context, Payload) (Result, error)

type Manager struct {
	client   *tgclient.Client      // Telegram API Client
	handlers map[string]Handler    // Map of registered command handlers.
	states   map[int64]HandlerFunc // Map of active user states.
}

func New(client *tgclient.Client) *Manager {
	return &Manager{
		client:   client,
		handlers: make(map[string]Handler),
		states:   make(map[int64]HandlerFunc),
	}
}

func (m *Manager) AddCommands(handlers ...Handler) {
	for _, handler := range handlers {
		prefix := "/" + handler.Prefix()
		m.handlers[prefix] = handler
		log.Printf("[INFO] %s registered \n", prefix)
	}
}

func (m *Manager) PublishCommands() {
	botCommands := make([]tgclient.BotCommand, 0, len(m.handlers))
	for _, handler := range m.handlers {
		botCommands = append(botCommands, tgclient.BotCommand{
			Command:     handler.Prefix(),
			Description: handler.Description(),
		})
	}

	cfg := tgclient.SetMyCommandsConfig{Commands: botCommands}
	_, err := m.client.SetMyCommands(context.Background(), cfg)
	if err != nil {
		log.Println("[ERROR]", err)
	}
}

func (m *Manager) Start() {
	cfg := tgclient.GetUpdatesConfig{
		Offset:         -1,
		Timeout:        60,
		AllowedUpdates: []string{"message", "callback_query"},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	updates := m.client.GetUpdatesChan(ctx, cfg, 100)
	for update := range updates {
		if update.Message != nil && update.Message.From != nil {
			pl := Payload{
				User:            update.Message.From,
				Chat:            update.Message.Chat,
				CallbackMessage: nil,
				Command:         update.Message.Text,
				ClearState:      true,
			}
			m.processUpdate(ctx, pl)
		} else if update.CallbackQuery != nil {
			pl := Payload{
				User:            update.CallbackQuery.From,
				Chat:            update.CallbackQuery.Message.Chat,
				CallbackMessage: update.CallbackQuery.Message,
				Command:         update.CallbackQuery.Data,
				ClearState:      false,
			}
			m.processUpdate(ctx, pl)
		}
	}
}

func (m *Manager) processUpdate(ctx context.Context, pl Payload) {
	fn, ok := m.getHandlerFunc(pl.User.ID, pl.Command)
	if !ok {
		return
	}

	res, err := fn(ctx, pl)
	if err != nil {
		log.Println("[ERROR]", err)
	}

	// If it is normal message, user can change and escape states
	// If it is callback event, user can change state only
	// If it is callback event, user can't clear his current state
	if res.State != nil {
		m.states[pl.User.ID] = res.State
	} else if pl.ClearState {
		delete(m.states, pl.User.ID)
	}

	// Update existing message with buttons
	if pl.CallbackMessage != nil && res.Keyboard != nil && res.Text != "" {
		_, err = m.client.EditMessageText(ctx, tgclient.EditMessageTextConfig{
			ChatId:      pl.Chat.ID,
			MessageID:   pl.CallbackMessage.MessageID,
			Text:        res.Text,
			ReplyMarkup: res.Keyboard,
		})
		if err != nil {
			log.Println("[ERROR]", err)
		}
	} else if res.Text != "" {
		_, err = m.client.SendMessage(ctx, tgclient.SendMessageConfig{
			ChatId:      pl.Chat.ID,
			Text:        res.Text,
			ReplyMarkup: res.Keyboard,
		})
		if err != nil {
			log.Println("[ERROR]", err)
		}
	}
}

func (m *Manager) getHandlerFunc(userID int64, text string) (HandlerFunc, bool) {
	if strings.HasPrefix(text, "/") { // New command
		prefix := strings.SplitN(text, " ", 2)[0]
		handler, ok := m.handlers[prefix]
		if ok {
			return handler.Execute, true
		}
	}
	fn, ok := m.states[userID] // Stateful command
	return fn, ok
}
