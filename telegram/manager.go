package telegram

import (
	"context"
	"log"
	"strings"

	"mr-weasel/commands"
)

type Manager struct {
	client   *Client                        // Telegram API Client
	handlers map[string]commands.Handler    // Map of registered command handlers.
	states   map[int64]commands.HandlerFunc // Map of active user states.
}

func NewManager(client *Client) *Manager {
	return &Manager{
		client:   client,
		handlers: make(map[string]commands.Handler),
		states:   make(map[int64]commands.HandlerFunc),
	}
}

func (m *Manager) AddCommands(handlers ...commands.Handler) {
	for _, handler := range handlers {
		prefix := handler.Prefix()
		m.handlers[prefix] = handler
		log.Printf("[INFO] %s registered \n", prefix)
	}
}

func (m *Manager) PublishCommands() {
	botCommands := make([]BotCommand, 0, len(m.handlers))
	for _, handler := range m.handlers {
		botCommands = append(botCommands, BotCommand{
			Command:     handler.Prefix(),
			Description: handler.Description(),
		})
	}

	cfg := SetMyCommandsConfig{Commands: botCommands}
	_, err := m.client.SetMyCommands(context.Background(), cfg)
	if err != nil {
		log.Println("[ERROR]", err)
	}
}

func (m *Manager) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := GetUpdatesConfig{
		Offset:         -1,
		Timeout:        60,
		AllowedUpdates: []string{"message", "callback_query"},
	}
	updates := m.client.GetUpdatesChan(ctx, cfg, 100)
	for update := range updates {
		m.processUpdate(ctx, update)
	}
}

func (m *Manager) processUpdate(ctx context.Context, update Update) {
	if update.Message != nil && update.Message.From != nil {
		m.processMessage(ctx, *update.Message)
	} else if update.CallbackQuery != nil {
		m.processCallbackQuery(ctx, *update.CallbackQuery)
	}
}

func (m *Manager) processMessage(ctx context.Context, message Message) {
	pl := commands.Payload{UserID: message.From.ID, Command: message.Text}
	res, err := m.executeCommand(ctx, pl)
	if err != nil {
		log.Println("[ERROR]", err)
	}

	// If it is normal message, user can change and escape states
	m.setState(pl.UserID, res.State, true)

	// Skip if there is no response text
	if res.Text == "" {
		return
	}

	botMessage, err := m.client.SendMessage(ctx, SendMessageConfig{
		ChatID:      message.Chat.ID,
		Text:        res.Text,
		ParseMode:   "HTML",
		ReplyMarkup: m.commandKeyboardToInlineMarkup(res.Keyboard),
	})
	if err != nil {
		log.Println("[ERROR]", err)
	}

	if res.ResultChan != nil {
		go func() {
			res = <-res.ResultChan
			_, err = m.client.SendMessage(ctx, SendMessageConfig{
				ChatID:           botMessage.Chat.ID,
				Text:             res.Text,
				ParseMode:        "HTML",
				ReplyMarkup:      m.commandKeyboardToInlineMarkup(res.Keyboard),
				ReplyToMessageId: botMessage.MessageID,
			})
			if err != nil {
				log.Println("[ERROR]", err)
			}
		}()
	}
}

func (m *Manager) processCallbackQuery(ctx context.Context, callbackQuery CallbackQuery) {
	pl := commands.Payload{UserID: callbackQuery.From.ID, Command: callbackQuery.Data}
	res, err := m.executeCommand(ctx, pl)
	if err != nil {
		log.Println("[ERROR]", err)
	}

	// If it is callback event, user can change state only
	// If it is callback event, user can't clear his current state
	m.setState(pl.UserID, res.State, false)

	_, err = m.client.AnswerCallbackQuery(ctx, AnswerCallbackQueryConfig{CallbackQueryID: callbackQuery.ID})
	if err != nil {
		log.Println("[ERROR]", err)
	}

	// Skip if there is no text and keyboard
	if res.Text == "" && res.Keyboard == nil {
		return
	}

	// If result text is empty, then use the original value
	if res.Text == "" {
		res.Text = callbackQuery.Message.Text
	}

	var botMessage Message

	if res.Keyboard != nil {
		botMessage, err = m.client.EditMessageText(ctx, EditMessageTextConfig{
			ChatID:      callbackQuery.Message.Chat.ID,
			MessageID:   callbackQuery.Message.MessageID,
			Text:        res.Text,
			ParseMode:   "HTML",
			ReplyMarkup: m.commandKeyboardToInlineMarkup(res.Keyboard),
		})
		if err != nil {
			log.Println("[ERROR]", err)
		}
	} else {
		botMessage, err = m.client.SendMessage(ctx, SendMessageConfig{
			ChatID:      callbackQuery.Message.Chat.ID,
			Text:        res.Text,
			ParseMode:   "HTML",
			ReplyMarkup: m.commandKeyboardToInlineMarkup(res.Keyboard),
		})
		if err != nil {
			log.Println("[ERROR]", err)
		}
	}

	if res.ResultChan != nil {
		go func() {
			res = <-res.ResultChan
			_, err = m.client.SendMessage(ctx, SendMessageConfig{
				ChatID:           botMessage.Chat.ID,
				Text:             res.Text,
				ParseMode:        "HTML",
				ReplyMarkup:      m.commandKeyboardToInlineMarkup(res.Keyboard),
				ReplyToMessageId: botMessage.MessageID,
			})
			if err != nil {
				log.Println("[ERROR]", err)
			}
		}()
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

func (m *Manager) commandKeyboardToInlineMarkup(keyboard [][]commands.Button) *InlineKeyboardMarkup {
	markup := &InlineKeyboardMarkup{InlineKeyboard: make([][]InlineKeyboardButton, len(keyboard))}
	for r, row := range keyboard {
		markup.InlineKeyboard[r] = make([]InlineKeyboardButton, len(row))
		for b, btn := range row {
			markup.InlineKeyboard[r][b] = InlineKeyboardButton{Text: btn.Label, CallbackData: btn.Data}
		}
	}
	return markup
}
