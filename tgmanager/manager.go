package tgmanager

import (
	"context"
	"log"
	"mr-weasel/commands"
	"mr-weasel/tgclient"
	"strings"
)

type Manager struct {
	client   *tgclient.Client               // Telegram API Client
	handlers map[string]commands.Handler    // Map of registered command handlers.
	states   map[int64]commands.HandlerFunc // Map of active user states.
}

func New(client *tgclient.Client) *Manager {
	return &Manager{
		client:   client,
		handlers: make(map[string]commands.Handler),
		states:   make(map[int64]commands.HandlerFunc),
	}
}

func (m *Manager) AddCommands(handlers ...commands.Handler) {
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
			pl := commands.Payload{UserID: update.Message.From.ID, Command: update.Message.Text}
			res, err := m.executeCommand(ctx, pl)
			if err != nil {
				log.Println("[ERROR]", err)
			} else {
				// If it is normal message, user can change and escape states
				m.setState(pl.UserID, res.State, true)
			}

			if res.Text != "" {
				_, err = m.client.SendMessage(ctx, tgclient.SendMessageConfig{
					ChatID:      update.Message.Chat.ID,
					Text:        res.Text,
					ReplyMarkup: res.Keyboard,
				})
				if err != nil {
					log.Println("[ERROR]", err)
				}
			}

		} else if update.CallbackQuery != nil {
			pl := commands.Payload{UserID: update.CallbackQuery.From.ID, Command: update.CallbackQuery.Data}
			res, err := m.executeCommand(ctx, pl)
			if err != nil {
				log.Println("[ERROR]", err)
			} else {
				// If it is callback event, user can change state only
				// If it is callback event, user can't clear his current state
				m.setState(pl.UserID, res.State, false)
			}

			_, err = m.client.AnswerCallbackQuery(ctx, tgclient.AnswerCallbackQueryConfig{
				CallbackQueryID: update.CallbackQuery.ID,
			})
			if err != nil {
				log.Println("[ERROR]", err)
			}

			if res.Text != "" && res.Keyboard != nil {
				_, err = m.client.EditMessageText(ctx, tgclient.EditMessageTextConfig{
					ChatID:      update.CallbackQuery.Message.Chat.ID,
					MessageID:   update.CallbackQuery.Message.MessageID,
					Text:        res.Text,
					ReplyMarkup: res.Keyboard,
				})
				if err != nil {
					log.Println("[ERROR]", err)
				}
			} else if res.Text != "" {
				_, err = m.client.SendMessage(ctx, tgclient.SendMessageConfig{
					ChatID:      update.CallbackQuery.Message.Chat.ID,
					Text:        res.Text,
					ReplyMarkup: res.Keyboard,
				})
				if err != nil {
					log.Println("[ERROR]", err)
				}
			}

		}
	}
}

func (m *Manager) getHandlerFunc(userID int64, text string) (commands.HandlerFunc, bool) {
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

func (m *Manager) executeCommand(ctx context.Context, pl commands.Payload) (commands.Result, error) {
	if fn, ok := m.getHandlerFunc(pl.UserID, pl.Command); ok {
		return fn(ctx, pl)
	}
	return commands.Result{}, nil
}

func (m *Manager) setState(userID int64, state commands.HandlerFunc, clear bool) {
	if state != nil {
		m.states[userID] = state
	} else if clear {
		delete(m.states, userID)
	}
}
