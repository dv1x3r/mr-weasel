package tgmanager

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"mr-weasel/commands"
	"mr-weasel/tgclient"
	"mr-weasel/utils"
)

type Manager struct {
	tgClient *tgclient.Client               // Telegram API Client
	handlers map[string]commands.Handler    // Map of registered command handlers.
	states   map[int64]commands.ExecuteFunc // Map of active user states.
	tokens   map[string]context.CancelFunc  // Map of cancellation tokens.
}

func NewManager(tgClient *tgclient.Client) *Manager {
	return &Manager{
		tgClient: tgClient,
		handlers: make(map[string]commands.Handler),
		states:   make(map[int64]commands.ExecuteFunc),
		tokens:   make(map[string]context.CancelFunc),
	}
}

func (m *Manager) AddCommands(handlers ...commands.Handler) []tgclient.BotCommand {
	botCommands := make([]tgclient.BotCommand, 0, len(handlers))

	for _, handler := range handlers {
		prefix := handler.Prefix()
		m.handlers[prefix] = handler

		botCommands = append(botCommands, tgclient.BotCommand{
			Command:     handler.Prefix(),
			Description: handler.Description(),
		})

		log.Printf("[INFO] Registered %s\n", prefix)
	}

	return botCommands
}

func (m *Manager) PublishCommands(botCommands []tgclient.BotCommand) {
	cfg := tgclient.SetMyCommandsConfig{Commands: botCommands}
	_, err := m.tgClient.SetMyCommands(context.Background(), cfg)
	if err != nil {
		log.Println("[ERROR]", err)
	}
}

func (m *Manager) Start(ctx context.Context) {
	cfg := tgclient.GetUpdatesConfig{
		Offset:         -1,
		Timeout:        60,
		AllowedUpdates: []string{"message", "callback_query"},
	}
	updates := m.tgClient.GetUpdatesChan(ctx, cfg, 100)
	for update := range updates {
		if update.Message != nil && update.Message.From != nil {
			m.onMessage(ctx, *update.Message)
		} else if update.CallbackQuery != nil {
			m.onCallbackQuery(ctx, *update.CallbackQuery)
			// Answer to the callback query just to dismiss "Loading..." prompt on the top
			_, err := m.tgClient.AnswerCallbackQuery(ctx, tgclient.AnswerCallbackQueryConfig{CallbackQueryID: update.CallbackQuery.ID})
			if err != nil {
				log.Println("[ERROR]", err)
			}
		}
	}
}

func (m *Manager) onMessage(ctx context.Context, message tgclient.Message) {
	const op = "telegram.Manager.processMessage"

	// command has /prefix@bot_username syntax
	message.Text = strings.TrimSuffix(message.Text, fmt.Sprintf("@%s", m.tgClient.Me.Username))

	execFn, ok := m.getExecuteFunc(message.From.ID, message.Text)
	if !ok {
		return
	}

	userName := "@" + message.From.Username
	if userName == "@" {
		userName = message.From.FirstName
	}

	pl := commands.Payload{
		UserID:     message.From.ID,
		UserName:   userName,
		IsPrivate:  message.Chat.Type == "private",
		Command:    message.Text,
		ResultChan: make(chan commands.Result),
	}

	if message.Audio != nil {
		fileURL, err := m.tgClient.GetFileURL(ctx, tgclient.GetFileConfig{FileID: message.Audio.FileID})
		if err != nil {
			log.Println("[ERROR]", utils.WrapIfErr(op, err))
			return
		}
		pl.FileURL = fileURL
		pl.Command = message.Audio.FileName
	} else if message.Voice != nil {
		fileURL, err := m.tgClient.GetFileURL(ctx, tgclient.GetFileConfig{FileID: message.Voice.FileID})
		if err != nil {
			log.Println("[ERROR]", utils.WrapIfErr(op, err))
			return
		}
		pl.FileURL = fileURL
		pl.Command = fmt.Sprintf("%s.oga", message.Voice.FileUniqueID)
	}

	if message.UserShared != nil {
		pl.Command = strconv.FormatInt(message.UserShared.UserID, 10)
	}

	go func() {
		ctx := context.WithValue(ctx, "contextID", fmt.Sprintf("%p", &pl))
		ctx, cancel := context.WithCancel(ctx)
		tokenKey := fmt.Sprintf("%d:%p", pl.UserID, &pl)
		m.tokens[tokenKey] = cancel

		defer close(pl.ResultChan)
		defer delete(m.tokens, tokenKey)

		execFn(ctx, pl)
	}()

	go m.processResults(ctx, pl, message)
}

func (m *Manager) onCallbackQuery(ctx context.Context, callbackQuery tgclient.CallbackQuery) {
	const op = "telegram.Manager.processCallbackQuery"

	// Check if chat user is message owner (for groups)
	if callbackQuery.Message.Chat.Type != "private" {
		if callbackQuery.Message.Entities[0].User.ID != callbackQuery.From.ID {
			return
		}
	}

	if strings.HasPrefix(callbackQuery.Data, commands.CmdCancel) {
		cancelFn, ok := m.getCancelFunc(callbackQuery.From.ID, callbackQuery.Data)
		if ok {
			cancelFn()
		}
		return
	}

	execFn, ok := m.getExecuteFunc(callbackQuery.From.ID, callbackQuery.Data)
	if !ok {
		return
	}

	userName := "@" + callbackQuery.From.Username
	if userName == "@" {
		userName = callbackQuery.From.FirstName
	}

	pl := commands.Payload{
		UserID:     callbackQuery.From.ID,
		UserName:   userName,
		IsPrivate:  callbackQuery.Message.Chat.Type == "private",
		Command:    callbackQuery.Data,
		ResultChan: make(chan commands.Result),
	}

	go func() {
		ctx := context.WithValue(ctx, "contextID", fmt.Sprintf("%p", &pl))
		ctx, cancel := context.WithCancel(ctx)
		tokenKey := fmt.Sprintf("%d:%p", pl.UserID, &pl)
		m.tokens[tokenKey] = cancel

		defer close(pl.ResultChan)
		defer delete(m.tokens, tokenKey)

		execFn(ctx, pl)
	}()

	go m.processResults(ctx, pl, *callbackQuery.Message)
}

func (m *Manager) processResults(ctx context.Context, pl commands.Payload, previousResponse tgclient.Message) {
	const op = "telegram.Manager.processResults"
	var err error

	for result := range pl.ResultChan {
		if result.Error != nil {
			log.Println("[ERROR]", utils.WrapIfErr(op, result.Error))
		}

		if result.Audio != nil {
			media := []tgclient.InputMedia{}

			keys := make([]string, 0, len(result.Audio))
			for k := range result.Audio {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, name := range keys {
				media = append(media, &tgclient.InputMediaAudio{Media: "attach://" + name})
			}

			_, err = m.tgClient.SendMediaGroup(ctx, tgclient.SendMediaGroupConfig{ChatID: previousResponse.Chat.ID, Media: media}, result.Audio)
			if err != nil {
				log.Println("[ERROR]", utils.WrapIfErr(op, err))
			}

		} else if result.InlineMarkup.InlineKeyboard != nil && previousResponse.ReplyMarkup != nil {
			// if both previous and new response contain an inline keyboard, then it is update

			// in case of update we can change states only, or if requested explicitly
			if result.State != nil {
				m.states[pl.UserID] = result.State
			} else if result.ClearState {
				delete(m.states, pl.UserID)
			}

			// in case of update, keep original text if not specified explicitly
			if result.Text == "" {
				result.Text = previousResponse.Text
			}

			if !pl.IsPrivate {
				result.Text = fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>\n\n%s", pl.UserID, pl.UserName, result.Text)
			}

			var replyMarkup *tgclient.InlineKeyboardMarkup
			if len(result.InlineMarkup.InlineKeyboard[0]) != 0 {
				replyMarkup = &result.InlineMarkup
			}

			previousResponse, err = m.tgClient.EditMessageText(ctx, tgclient.EditMessageTextConfig{
				ChatID:      previousResponse.Chat.ID,
				MessageID:   previousResponse.MessageID,
				Text:        result.Text,
				ParseMode:   "HTML",
				ReplyMarkup: replyMarkup,
			})
			if err != nil {
				log.Println("[ERROR]", utils.WrapIfErr(op, err))
			}

		} else if result.Text != "" {
			// otherwise it is just a new message

			// in case of new reponse message we can both change and escape states
			if result.State != nil {
				m.states[pl.UserID] = result.State
			} else {
				delete(m.states, pl.UserID)
			}

			var replyMarkup tgclient.ReplyMarkup
			if result.InlineMarkup.InlineKeyboard != nil {
				replyMarkup = result.InlineMarkup
			} else if result.ReplyMarkup.Keyboard != nil {
				replyMarkup = result.ReplyMarkup
			} else if result.RemoveMarkup.RemoveKeyboard {
				replyMarkup = result.RemoveMarkup
			}

			if !pl.IsPrivate {
				result.Text = fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>\n\n%s", pl.UserID, pl.UserName, result.Text)
			}

			previousResponse, err = m.tgClient.SendMessage(ctx, tgclient.SendMessageConfig{
				ChatID:      previousResponse.Chat.ID,
				Text:        result.Text,
				ParseMode:   "HTML",
				ReplyMarkup: replyMarkup,
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

func (m *Manager) getCancelFunc(userID int64, text string) (context.CancelFunc, bool) {
	split := strings.SplitN(text, " ", 2)
	if len(split) != 2 {
		return nil, false
	}

	tokenKey := fmt.Sprintf("%d:%s", userID, split[1])
	cancel, ok := m.tokens[tokenKey]
	if ok {
		log.Printf("[VERB] %d: %s\n", userID, text)
	}

	return cancel, ok
}
