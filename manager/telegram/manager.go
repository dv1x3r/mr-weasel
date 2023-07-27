package tgmanager

import (
	"context"
	"fmt"
	"log"
	"mr-weasel/client/telegram"
	"strings"
)

type Command struct {
	Prefix string
	Action string
	Text   string
}

type Payload struct {
	User    tgclient.User
	Command Command
}

type Result struct {
	Action   string
	Text     string
	Keyboard *tgclient.InlineKeyboardMarkup
}

func (res *Result) AddKeyboardButton(row int, text string, data string) {
	if res.Keyboard == nil {
		res.Keyboard = &tgclient.InlineKeyboardMarkup{}
		res.Keyboard.InlineKeyboard = [][]tgclient.InlineKeyboardButton{}
	}

	for i := len(res.Keyboard.InlineKeyboard); i < row+1; i++ {
		res.Keyboard.InlineKeyboard = append(res.Keyboard.InlineKeyboard, []tgclient.InlineKeyboardButton{})
	}

	res.Keyboard.InlineKeyboard[row] = append(res.Keyboard.InlineKeyboard[row], tgclient.InlineKeyboardButton{Text: text, CallbackData: data})
}

type Handler interface {
	Prefix() string
	Description() string
	ExecuteTelegram(Payload) (Result, error)
}

type Manager struct {
	client   *tgclient.Client   // Telegram API Client
	debug    bool               // Enable debug output.
	handlers map[string]Handler // Map of all registered command handlers.
	states   map[int64]Command  // Map of all active user states (active commands).
}

func New(client *tgclient.Client, debug bool) *Manager {
	return &Manager{
		client:   client,
		debug:    debug,
		handlers: make(map[string]Handler),
		states:   make(map[int64]Command),
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
			res, err := m.execute(*update.Message.From, update.Message.Text)
			if err != nil {
				log.Println("[ERROR] Message execution:", err)
				continue
			}

			_, err = m.client.SendMessage(ctx, tgclient.SendMessageConfig{
				ChatId:      update.Message.Chat.ID,
				Text:        res.Text,
				ReplyMarkup: res.Keyboard,
			})
			if err != nil {
				log.Println("[ERROR] Sending a text response:", err)
			}

		} else if update.CallbackQuery != nil {
			res, err := m.execute(*update.CallbackQuery.From, update.CallbackQuery.Data)
			if err != nil {
				log.Println("[ERROR] Callback execution:", err)
				continue
			}

			if res.Keyboard == nil {
				_, err = m.client.SendMessage(ctx, tgclient.SendMessageConfig{
					ChatId: update.CallbackQuery.Message.Chat.ID,
					Text:   res.Text,
				})
				if err != nil {
					log.Println("[ERROR] Sending a callback text response:", err)
				}
			} else {
				_, err = m.client.EditMessageText(ctx, tgclient.EditMessageTextConfig{
					ChatId:      update.CallbackQuery.Message.Chat.ID,
					MessageID:   update.CallbackQuery.Message.MessageID,
					Text:        res.Text,
					ReplyMarkup: res.Keyboard,
				})
				if err != nil {
					log.Println("[ERROR] Update a callback text response:", err)
				}
			}

		}
	}
}

func (m *Manager) execute(user tgclient.User, input string) (Result, error) {
	command, handler := m.getCommandHandler(user.ID, input)
	if handler == nil {
		return Result{}, nil
	}

	res, err := handler.ExecuteTelegram(Payload{User: user, Command: command})
	if err != nil {
		return Result{}, err
	}

	if res.Action != "" {
		m.states[user.ID] = Command{Prefix: command.Prefix, Action: res.Action}
	} else {
		delete(m.states, user.ID)
	}

	if res.Text == "" {
		return res, fmt.Errorf("command [%v] returned empty text", command.Prefix)
	}

	return res, nil
}

func (m *Manager) getCommandHandler(userID int64, text string) (Command, Handler) {
	command, isCommand := parseCommand(text)         // Split message by /prefix:action text
	handler, isHandler := m.handlers[command.Prefix] // Get the command handler
	if isCommand && isHandler {                      // Execute a new command
		return command, handler
	}

	state, isState := m.states[userID]
	if !isState {
		return Command{}, nil
	}

	handler = m.handlers[state.Prefix]
	state.Text = text // User answered to the bot's question in the current state
	return state, handler
}

func parseCommand(text string) (Command, bool) {
	safeGet := func(arr []string, i int) string {
		if len(arr)-1 >= i {
			return arr[i]
		}
		return ""
	}

	if !strings.HasPrefix(text, "/") {
		return Command{}, false
	}

	s := strings.SplitN(text, " ", 2)
	cmd, text := safeGet(s, 0), safeGet(s, 1)

	s = strings.Split(cmd, ":")
	prefix, action := safeGet(s, 0), safeGet(s, 1)

	return Command{Prefix: prefix, Action: action, Text: text}, true
}
