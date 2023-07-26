package tgmanager

import (
	"context"
	"log"
	"mr-weasel/client/telegram"
	"strings"
)

type User struct {
	UserID int64
}

type Payload struct {
	Prefix string
	Action string
	Text   string
}

type Result struct {
	Text   string
	Action string
	Markup *tgclient.InlineKeyboardMarkup
}

type State struct {
	Prefix string
	Action string
}

type Handler interface {
	Prefix() string
	Description() string
	ExecuteTelegram(User, Payload) (Result, error)
}

type Manager struct {
	client   *tgclient.Client   // Telegram API Client
	debug    bool               // Enable debug output.
	handlers map[string]Handler // Map of all registered command handlers.
	states   map[int64]State    // Map of all active user states (active commands).
}

func New(client *tgclient.Client, debug bool) *Manager {
	return &Manager{
		client:   client,
		debug:    debug,
		handlers: make(map[string]Handler),
		states:   make(map[int64]State),
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

func parsePayload(text string) Payload {
	safeGet := func(arr []string, i int) string {
		if len(arr)-1 >= i {
			return arr[i]
		}
		return ""
	}

	s := strings.SplitN(text, " ", 2)
	command, text := safeGet(s, 0), safeGet(s, 1)

	s = strings.Split(command, ":")
	prefix, action := safeGet(s, 0), safeGet(s, 1)

	return Payload{Prefix: prefix, Action: action, Text: text}
}

func (m *Manager) getCommand(message *tgclient.Message) (Payload, Handler) {
	payload := parsePayload(message.Text)     // Split message by /prefix:action text
	handler, ok := m.handlers[payload.Prefix] // Get the command handler (if exists)
	if ok {
		return payload, handler
	}

	state, ok := m.states[message.From.ID] // Check if user has an active state
	if ok {
		payload = Payload{Prefix: state.Prefix, Action: state.Action, Text: message.Text}
		handler = m.handlers[state.Prefix] // Get the command handler for that state
		return payload, handler
	}

	return Payload{}, nil
}

func (m *Manager) updateState(userID int64, prefix string, action string) {
	if action != "" {
		m.states[userID] = State{Prefix: prefix, Action: action}
	} else {
		delete(m.states, userID)
	}
}

func (m *Manager) processMessage(msg *tgclient.Message) {
	payload, handler := m.getCommand(msg)
	if handler == nil {
		if m.debug {
			log.Println("[DEBUG] Handler not found:", msg.Text)
		}
		return
	}

	res, err := handler.ExecuteTelegram(User{UserID: msg.From.ID}, payload)
	if err != nil {
		log.Printf("[ERROR] %+v %s \n", payload, err)
		return
	}

	m.updateState(msg.From.ID, payload.Prefix, res.Action)

	log.Printf("[INFO] %+v succeeded \n", payload)

	if res.Text != "" {
		_, err = m.client.SendMessage(context.Background(), tgclient.SendMessageConfig{
			ChatId: msg.Chat.ID,
			Text:   res.Text,
		})
		if err != nil {
			log.Println("[ERROR] Sending a response:", err)
		}
	}
}

func (m *Manager) processCallbackQuery(cq *tgclient.CallbackQuery) {

}
