package tgmanager

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"mr-weasel/client/telegram"
	"strings"
)

var (
	ErrCommandNotFound = errors.New("command not found")
	ErrCommandFailed   = errors.New("command failed")
	ErrCommandEmpty    = errors.New("command returned empty text")
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
	res.Keyboard.InlineKeyboard = append(res.Keyboard.InlineKeyboard,
		[]tgclient.InlineKeyboardButton{})
}

type Handler interface {
	Prefix() string
	Description() string
	Execute(context.Context, *sqlx.DB, Payload) (Result, error)
}

type HandlerFunc = func(context.Context, *sqlx.DB, Payload) (Result, error)

type Manager struct {
	client   *tgclient.Client      // Telegram API Client
	db       *sqlx.DB              // Database storage
	handlers map[string]Handler    // Map of registered command handlers.
	states   map[int64]HandlerFunc // Map of active user states.
}

func New(client *tgclient.Client, db *sqlx.DB, debug bool) *Manager {
	return &Manager{
		client:   client,
		db:       db,
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
			command := update.Message.Text

			res, err := m.executeCommand(ctx, user, command)
			if err != nil {
				log.Println("[ERROR]", err)
				continue
			}

			// If it is normal message, user can change and escape states
			if res.State != nil {
				m.states[user.ID] = res.State
			} else {
				delete(m.states, user.ID)
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
			user := update.CallbackQuery.From
			command := update.CallbackQuery.Data

			res, err := m.executeCommand(ctx, user, command)
			if err != nil {
				log.Println("[ERROR]", err)
				continue
			}

			// If it is callback event, user can change state only
			// User should not remove the current state
			if res.State != nil {
				m.states[user.ID] = res.State
			}

			if res.Keyboard == nil {
				_, err = m.client.SendMessage(ctx, tgclient.SendMessageConfig{
					ChatId: update.CallbackQuery.Message.Chat.ID,
					Text:   res.Text,
				})
				if err != nil {
					log.Println("[ERROR] Callback send message:", err)
				}
			} else {
				_, err = m.client.EditMessageText(ctx, tgclient.EditMessageTextConfig{
					ChatId:      update.CallbackQuery.Message.Chat.ID,
					MessageID:   update.CallbackQuery.Message.MessageID,
					Text:        res.Text,
					ReplyMarkup: res.Keyboard,
				})
				if err != nil {
					log.Println("[ERROR] Callback edit message:", err)
				}
			}
		}
	}
}

func (m *Manager) executeCommand(ctx context.Context, user *tgclient.User, command string) (Result, error) {
	fn, ok := m.getHandlerFunc(user.ID, command)
	if !ok {
		return Result{}, fmt.Errorf("executeCommand %s %w", command, ErrCommandNotFound)
	}

	res, err := fn(ctx, m.db, Payload{User: *user, Command: command})
	if err != nil {
		return res, fmt.Errorf("executeCommand %s %w %w", command, err, ErrCommandFailed)
	}

	if res.Text == "" {
		return Result{}, fmt.Errorf("executeCommand %s %w", command, ErrCommandEmpty)
	}

	return res, nil
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
