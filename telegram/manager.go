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
		if update.Message != nil && update.Message.From != nil {
			m.processMessage(ctx, *update.Message)
		} else if update.CallbackQuery != nil {
			m.processCallbackQuery(ctx, *update.CallbackQuery)
			// Answer to the callback query just to dismiss "Loading..." prompt on the top
			_, err := m.tgClient.AnswerCallbackQuery(ctx, AnswerCallbackQueryConfig{CallbackQueryID: update.CallbackQuery.ID})
			if err != nil {
				log.Println("[ERROR]", err)
			}
		}
	}
}

func (m *Manager) processMessage(ctx context.Context, message Message) {
	const op = "telegram.Manager.processMessage"
	fn, ok := m.getExecuteFunc(message.From.ID, message.Text)
	if !ok {
		return
	}

	var blobPayload *utils.BlobPayload
	if message.Audio != nil {
		URL, err := m.tgClient.GetFileURL(ctx, GetFileConfig{FileID: message.Audio.FileID})
		if err != nil {
			log.Println("[ERROR]", utils.WrapIfErr(op, err))
			return
		}
		blobPayload = &utils.BlobPayload{
			FileID:   message.Audio.FileID,
			FileName: message.Audio.FileName,
			URL:      URL,
		}
	}

	pl := commands.Payload{
		UserID:      message.From.ID,
		Command:     message.Text,
		BlobPayload: blobPayload,
		ResultChan:  make(chan commands.Result),
	}

	go func() {
		defer close(pl.ResultChan)
		fn(ctx, pl)
	}()

	go m.processResults(ctx, pl, message)
}

func (m *Manager) processCallbackQuery(ctx context.Context, callbackQuery CallbackQuery) {
	const op = "telegram.Manager.processCallbackQuery"
	fn, ok := m.getExecuteFunc(callbackQuery.From.ID, callbackQuery.Data)
	if !ok {
		return
	}

	pl := commands.Payload{
		UserID:     callbackQuery.From.ID,
		Command:    callbackQuery.Data,
		ResultChan: make(chan commands.Result),
	}

	go func() {
		defer close(pl.ResultChan)
		fn(ctx, pl)
	}()

	go m.processResults(ctx, pl, *callbackQuery.Message)
}

func (m *Manager) processResults(ctx context.Context, pl commands.Payload, previousResponse Message) {
	const op = "telegram.Manager.processResults"
	var err error

	for result := range pl.ResultChan {
		if result.Error != nil {
			log.Println("[ERROR]", utils.WrapIfErr(op, result.Error))
		}

		// if both previous and new response contain a keyboard, then it is update
		if result.Keyboard != nil && previousResponse.ReplyMarkup != nil {
			// in case of update we can both only change states
			if result.State != nil {
				m.states[pl.UserID] = result.State
			}

			// in case of update, keep original text if not specified explicitly
			if result.Text == "" {
				result.Text = previousResponse.Text
			}

			previousResponse, err = m.tgClient.EditMessageText(ctx, EditMessageTextConfig{
				ChatID:      previousResponse.Chat.ID,
				MessageID:   previousResponse.MessageID,
				Text:        result.Text,
				ParseMode:   "HTML",
				ReplyMarkup: m.commandKeyboardToInlineMarkup(result.Keyboard),
			})
			if err != nil {
				log.Println("[ERROR]", utils.WrapIfErr(op, err))
			}

		} else if result.Text != "" {
			// in case of new reponse message we can both change and escape states
			if result.State != nil {
				m.states[pl.UserID] = result.State
			} else {
				delete(m.states, pl.UserID)
			}

			previousResponse, err = m.tgClient.SendMessage(ctx, SendMessageConfig{
				ChatID:      previousResponse.Chat.ID,
				Text:        result.Text,
				ParseMode:   "HTML",
				ReplyMarkup: m.commandKeyboardToInlineMarkup(result.Keyboard),
			})
			if err != nil {
				log.Println("[ERROR]", utils.WrapIfErr(op, err))
			}
		}

	}
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
