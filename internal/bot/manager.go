package bot

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"mr-weasel/internal/commands"
	"mr-weasel/internal/lib/telegram"
	"mr-weasel/internal/lib/wrap"
)

type Manager struct {
	client   *telegram.Client               // telegram api client
	handlers map[string]commands.Handler    // registered command handlers
	states   map[int64]commands.ExecuteFunc // active user states
	tokens   map[string]context.CancelFunc  // cancellation tokens
}

func NewManager(client *telegram.Client) *Manager {
	return &Manager{
		client:   client,
		handlers: map[string]commands.Handler{},
		states:   map[int64]commands.ExecuteFunc{},
		tokens:   map[string]context.CancelFunc{},
	}
}

func (m *Manager) AddCommands(handlers ...commands.Handler) []telegram.BotCommand {
	botCommands := make([]telegram.BotCommand, 0, len(handlers))

	for _, handler := range handlers {
		prefix := handler.Prefix()
		m.handlers[prefix] = handler

		botCommands = append(botCommands, telegram.BotCommand{
			Command:     handler.Prefix(),
			Description: handler.Description(),
		})

		log.Printf("[INFO] Registered %s\n", prefix)
	}

	return botCommands
}

func (m *Manager) PublishCommands(botCommands []telegram.BotCommand) {
	cfg := telegram.SetMyCommandsConfig{Commands: botCommands}
	if _, err := m.client.SetMyCommands(context.Background(), cfg); err != nil {
		log.Println("[ERROR]", err)
	}
}

func (m *Manager) Start(ctx context.Context) {
	cfg := telegram.GetUpdatesConfig{
		Offset:         -1,
		Timeout:        60,
		AllowedUpdates: []string{"message", "callback_query"},
	}
	updates := m.client.GetUpdatesChan(ctx, cfg, 100)
	for update := range updates {
		if update.Message != nil && update.Message.From != nil {
			m.onMessage(ctx, *update.Message)
		} else if update.CallbackQuery != nil {
			m.onCallbackQuery(ctx, *update.CallbackQuery)
			// Answer to the callback query just to dismiss "Loading..." prompt on the top
			if _, err := m.client.AnswerCallbackQuery(ctx, telegram.AnswerCallbackQueryConfig{CallbackQueryID: update.CallbackQuery.ID}); err != nil {
				log.Println("[ERROR]", err)
			}
		}
	}
}

func (m *Manager) onMessage(ctx context.Context, message telegram.Message) {
	const op = "bot.Manager.processMessage"

	// command has /prefix@bot_username syntax
	message.Text = strings.TrimSuffix(message.Text, fmt.Sprintf("@%s", m.client.Me.Username))

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
		fileURL, err := m.client.GetFileURL(ctx, telegram.GetFileConfig{FileID: message.Audio.FileID})
		if err != nil {
			log.Println("[ERROR]", wrap.IfErr(op, err))
			return
		}
		pl.FileURL = fileURL
		pl.Command = message.Audio.FileName
	} else if message.Voice != nil {
		fileURL, err := m.client.GetFileURL(ctx, telegram.GetFileConfig{FileID: message.Voice.FileID})
		if err != nil {
			log.Println("[ERROR]", wrap.IfErr(op, err))
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

func (m *Manager) onCallbackQuery(ctx context.Context, callbackQuery telegram.CallbackQuery) {
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

func (m *Manager) processResults(ctx context.Context, pl commands.Payload, previousResponse telegram.Message) {
	const op = "bot.Manager.processResults"
	var err error

	for result := range pl.ResultChan {
		if result.Error != nil {
			log.Println("[ERROR]", wrap.IfErr(op, result.Error))
		}

		if previousResponse.Chat == nil {
			log.Println("[WARN]", "Chat no longer exists, bot has been kicked!")
			return
		}

		if result.Audio != nil {
			media := []telegram.InputMedia{}

			keys := make([]string, 0, len(result.Audio))
			for k := range result.Audio {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, name := range keys {
				media = append(media, &telegram.InputMediaAudio{Media: "attach://" + name})
			}

			_, err = m.client.SendMediaGroup(ctx, telegram.SendMediaGroupConfig{ChatID: previousResponse.Chat.ID, Media: media}, result.Audio)
			if err != nil {
				log.Println("[ERROR]", wrap.IfErr(op, err))
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

			if !pl.IsPrivate && result.Text != previousResponse.Text {
				result.Text = fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>\n\n%s", pl.UserID, pl.UserName, result.Text)
			}

			var replyMarkup *telegram.InlineKeyboardMarkup
			if len(result.InlineMarkup.InlineKeyboard[0]) != 0 {
				replyMarkup = &result.InlineMarkup
			}

			previousResponse, err = m.client.EditMessageText(ctx, telegram.EditMessageTextConfig{
				ChatID:      previousResponse.Chat.ID,
				MessageID:   previousResponse.MessageID,
				Text:        result.Text,
				ParseMode:   "HTML",
				ReplyMarkup: replyMarkup,
			})
			if err != nil {
				log.Println("[ERROR]", wrap.IfErr(op, err))
			}

		} else if result.Text != "" {
			// otherwise it is just a new message

			// in case of new reponse message we can both change and escape states
			if result.State != nil {
				m.states[pl.UserID] = result.State
			} else {
				delete(m.states, pl.UserID)
			}

			var replyMarkup telegram.ReplyMarkup
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

			previousResponse, err = m.client.SendMessage(ctx, telegram.SendMessageConfig{
				ChatID:      previousResponse.Chat.ID,
				Text:        result.Text,
				ParseMode:   "HTML",
				ReplyMarkup: replyMarkup,
			})
			if err != nil {
				log.Println("[ERROR]", wrap.IfErr(op, err))
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
		log.Printf("[VERB] %d: %s\n", userID, runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name())
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
