package telegram

import (
	"context"
	"log"
	"strings"

	"mr-weasel/commands"
	"mr-weasel/utils"
)

type Manager struct {
	tgClient *Client                        // Telegram API Client
	handlers map[string]commands.Handler    // Map of registered command handlers.
	states   map[int64]commands.ExecuteFunc // Map of active user states.
}

func NewManager(tgClient *Client) *Manager {
	return &Manager{
		tgClient: tgClient,
		handlers: make(map[string]commands.Handler),
		states:   make(map[int64]commands.ExecuteFunc),
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
	const op = "telegram.Manager.PublishCommands"
	botCommands := make([]BotCommand, 0, len(m.handlers))
	for _, handler := range m.handlers {
		botCommands = append(botCommands, BotCommand{
			Command:     handler.Prefix(),
			Description: handler.Description(),
		})
	}

	cfg := SetMyCommandsConfig{Commands: botCommands}
	_, err := m.tgClient.SetMyCommands(context.Background(), cfg)
	if err != nil {
		log.Println("[ERROR]", utils.WrapIfErr(op, err))
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
	updates := m.tgClient.GetUpdatesChan(ctx, cfg, 100)
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
	const op = "telegram.Manager.processMessage"
	fn, ok := m.getExecuteFunc(message.From.ID, message.Text)
	if !ok {
		return
	}

	var fileURL string
	var err error

	if message.Audio != nil {
		fileURL, err = m.tgClient.GetFileURL(ctx, GetFileConfig{FileID: message.Audio.FileID})
		if err != nil {
			log.Println("[ERROR]", utils.WrapIfErr(op, err))
			return
		}
	}

	pl := commands.Payload{
		UserID:     message.From.ID,
		Command:    message.Text,
		FileURL:    fileURL,
		ResultChan: make(chan commands.Result),
	}

	go func() {
		defer close(pl.ResultChan)
		fn(ctx, pl)
	}()

	go func() {
		var previousResponse Message
		var err error

		for result := range pl.ResultChan {

			if result.Error != nil {
				log.Println("[ERROR]", utils.WrapIfErr(op, result.Error))
			}

			if result.State != nil {
				m.states[pl.UserID] = result.State
			} else if previousResponse.MessageID == 0 {
				// Only the root response can clear the state
				delete(m.states, pl.UserID)
			}

			if result.Text == "" {
				continue
			}

			previousResponse, err = m.tgClient.SendMessage(ctx, SendMessageConfig{
				ChatID:           message.Chat.ID,
				Text:             result.Text,
				ParseMode:        "HTML",
				ReplyMarkup:      m.commandKeyboardToInlineMarkup(result.Keyboard),
				ReplyToMessageId: previousResponse.MessageID,
			})
			if err != nil {
				log.Println("[ERROR]", utils.WrapIfErr(op, err))
			}

		}
	}()
}

func (m *Manager) processCallbackQuery(ctx context.Context, callbackQuery CallbackQuery) {
	const op = "telegram.Manager.processCallbackQuery"
	pl := commands.Payload{
		UserID:     callbackQuery.From.ID,
		Command:    callbackQuery.Data,
		ResultChan: make(chan commands.Result),
	}

	fn, ok := m.getExecuteFunc(pl.UserID, pl.Command)
	if !ok {
		return
	}

	// Answer to the callback query just to dismiss "Loading..." prompt on the top
	_, err := m.tgClient.AnswerCallbackQuery(ctx, AnswerCallbackQueryConfig{CallbackQueryID: callbackQuery.ID})
	if err != nil {
		log.Println("[ERROR]", utils.WrapIfErr(op, err))
	}

	go func() {
		defer close(pl.ResultChan)
		fn(ctx, pl)
	}()

	go func() {
		var previousResponse Message

		for result := range pl.ResultChan {

			if result.Error != nil {
				log.Println("[ERROR]", utils.WrapIfErr(op, result.Error))
			}

			if result.State != nil {
				m.states[pl.UserID] = result.State
			}

			// Skip if there is no text and keyboard
			if result.Text == "" && result.Keyboard == nil {
				continue
			}

			// If result text is empty, then use the original value
			// (useful for calendar script)
			if result.Text == "" {
				result.Text = callbackQuery.Message.Text
			}

			if previousResponse.MessageID == 0 {
				if result.Keyboard != nil {
					// Root response with keyboard changes callback message
					previousResponse, err = m.tgClient.EditMessageText(ctx, EditMessageTextConfig{
						ChatID:      callbackQuery.Message.Chat.ID,
						MessageID:   callbackQuery.Message.MessageID,
						Text:        result.Text,
						ParseMode:   "HTML",
						ReplyMarkup: m.commandKeyboardToInlineMarkup(result.Keyboard),
					})
					if err != nil {
						log.Println("[ERROR]", utils.WrapIfErr(op, err))
					}
				} else {
					// Root response with no keyboard spawns new message
					previousResponse, err = m.tgClient.SendMessage(ctx, SendMessageConfig{
						ChatID:    callbackQuery.Message.Chat.ID,
						Text:      result.Text,
						ParseMode: "HTML",
					})
					if err != nil {
						log.Println("[ERROR]", utils.WrapIfErr(op, err))
					}
				}
			} else {
				// Each channel update response spawns new message (background processing)
				previousResponse, err = m.tgClient.SendMessage(ctx, SendMessageConfig{
					ChatID:      callbackQuery.Message.Chat.ID,
					Text:        result.Text,
					ParseMode:   "HTML",
					ReplyMarkup: m.commandKeyboardToInlineMarkup(result.Keyboard),
				})
				if err != nil {
					log.Println("[ERROR]", utils.WrapIfErr(op, err))
				}
			}

		}
	}()

}

func (m *Manager) getExecuteFunc(userID int64, text string) (commands.ExecuteFunc, bool) {
	if strings.HasPrefix(text, "/") { // New command
		prefix := strings.SplitN(text, " ", 2)[0]
		handler, ok := m.handlers[prefix]
		if ok {
			log.Printf("[VERB] %d: %s\n", userID, text)
			return handler.Execute, true
		}
	}
	fn, ok := m.states[userID] // Stateful command
	if ok {
		log.Printf("[VERB] %d: %s\n", userID, utils.GetFunctionName(fn))
	}
	return fn, ok
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
