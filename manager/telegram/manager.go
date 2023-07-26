package tgmanager

import (
	"context"
	"log"
	"mr-weasel/client/telegram"
	"strings"
)

type User = tgclient.User
type InlineKeyboardMarkup = tgclient.InlineKeyboardMarkup
type InlineKeyboardButton = tgclient.InlineKeyboardButton

type Command struct {
	Prefix string
	Action string
	Text   string
}

type Payload struct {
	User    User
	Command Command
}

type Result struct {
	Text     string
	Action   string
	Keyboard *InlineKeyboardMarkup
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
		log.Println("[ERROR] UploadCommands:", err)
	}

	log.Println("[INFO] UploadCommands:", res)
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
		if update.Message != nil {
			m.processMessage(update.Message)
		}
		if update.CallbackQuery != nil {
			m.processCallbackQuery(update.CallbackQuery)
		}
	}
}

func parseCommand(text string) Command {
	safeGet := func(arr []string, i int) string {
		if len(arr)-1 >= i {
			return arr[i]
		}
		return ""
	}

	s := strings.SplitN(text, " ", 2)
	cmd, text := safeGet(s, 0), safeGet(s, 1)

	s = strings.Split(cmd, ":")
	prefix, action := safeGet(s, 0), safeGet(s, 1)

	return Command{Prefix: prefix, Action: action, Text: text}
}

func (m *Manager) getCommandHandler(userID int64, text string) (Command, Handler) {
	command := parseCommand(text)             // Split message by /prefix:action text
	handler, ok := m.handlers[command.Prefix] // Get the command handler (if exists)
	if ok {
		return command, handler
	}

	state, ok := m.states[userID] // Check if user has an active state
	if ok {
		state.Text = text
		handler = m.handlers[state.Prefix] // Get the command handler for that state
		return state, handler
	}

	return Command{}, nil
}

func (m *Manager) processMessage(msg *tgclient.Message) {
	command, handler := m.getCommandHandler(msg.From.ID, msg.Text)
	if handler == nil {
		if m.debug {
			log.Println("[DEBUG] Handler not found:", msg.Text)
		}
		return
	}

	res, err := handler.ExecuteTelegram(Payload{User: *msg.From, Command: command})
	if err != nil {
		log.Printf("[ERROR] %+v %s \n", command, err)
		return
	}

	if res.Action != "" {
		m.states[msg.From.ID] = Command{Prefix: command.Prefix, Action: res.Action}
	} else {
		delete(m.states, msg.From.ID)
	}

	log.Printf("[INFO] %+v succeeded \n", command)

	if res.Text == "" {
		return
	}

	_, err = m.client.SendMessage(context.Background(), tgclient.SendMessageConfig{
		ChatId:      msg.Chat.ID,
		Text:        res.Text,
		ReplyMarkup: res.Keyboard,
	})
	if err != nil {
		log.Println("[ERROR] Sending a response:", err)
	}
}

func (m *Manager) processCallbackQuery(cq *tgclient.CallbackQuery) {

}
