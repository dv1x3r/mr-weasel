package tgmanager

import (
	"context"
	"log"
	"mr-weasel/client/telegram"
	"strings"
)

type Payload struct {
	User    tgclient.User
	Command string
}

type Result struct {
	Text     string
	State    HandlerFunc
	Keyboard *tgclient.InlineKeyboardMarkup
}

type Handler interface {
	Prefix() string
	Description() string
	Execute(Payload) (Result, error)
}

type HandlerFunc = func(Payload) (Result, error)

type Manager struct {
	client   *tgclient.Client      // Telegram API Client
	handlers map[string]Handler    // Map of registered command handlers.
	states   map[int64]HandlerFunc // Map of active user states.
}

func New(client *tgclient.Client, debug bool) *Manager {
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
	res, err := m.client.SetMyCommands(context.Background(), cfg)
	if err != nil {
		log.Println("[ERROR] Publishing commands:", err)
	}

	log.Println("[INFO] Publishing commands:", res)
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
			user := update.Message.From
			text := update.Message.Text

			fn, ok := m.getHandlerFunc(user.ID, text)
			if !ok {
				log.Printf("[WARN] %s handler not found \n", text)
				continue
			}

			res, err := fn(Payload{User: *user, Command: text})
			if err != nil {
				log.Printf("[ERROR] %s %s \n", text, err)
				continue
			}

			if res.State != nil {
				m.states[user.ID] = res.State
			} else {
				delete(m.states, user.ID)
			}

			if res.Text == "" {
				log.Printf("[WARN] %s empty response text \n", text)
				continue
			}

			_, err = m.client.SendMessage(ctx, tgclient.SendMessageConfig{
				ChatId:      update.Message.Chat.ID,
				Text:        res.Text,
				ReplyMarkup: res.Keyboard,
			})
			if err != nil {
				log.Println("[ERROR] Send message error:", err)
			}
		} else if update.CallbackQuery != nil {
			// res, err := m.execute(*update.CallbackQuery.From, update.CallbackQuery.Data)
			// if err != nil {
			// 	log.Println("[ERROR] Callback execution:", err)
			// 	continue
			// }

			// if res.Keyboard == nil {
			// 	_, err = m.client.SendMessage(ctx, tgclient.SendMessageConfig{
			// 		ChatId: update.CallbackQuery.Message.Chat.ID,
			// 		Text:   res.Text,
			// 	})
			// 	if err != nil {
			// 		log.Println("[ERROR] Sending a callback text response:", err)
			// 	}
			// } else {
			// 	_, err = m.client.EditMessageText(ctx, tgclient.EditMessageTextConfig{
			// 		ChatId:      update.CallbackQuery.Message.Chat.ID,
			// 		MessageID:   update.CallbackQuery.Message.MessageID,
			// 		Text:        res.Text,
			// 		ReplyMarkup: res.Keyboard,
			// 	})
			// 	if err != nil {
			// 		log.Println("[ERROR] Update a callback text response:", err)
			// 	}
			// }

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
