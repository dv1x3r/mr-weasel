package tgmanager

import (
	"context"
	"log"
	"mr-weasel/client/telegram"
	"strings"
)

type Input struct {
	User tgclient.User // User who initiated command.
	Text string        // Original message text.
}

type Output struct {
	Text  string // New message text.
	State string // Set user state.
}

type Handler interface {
	Prefix() string
	Description() string
	ExecuteTelegram(Input) (Output, error)
}

type HandlerFunc = func(Input) (Output, error)

type Manager struct {
	client   *tgclient.Client       // Telegram API Client
	commands map[string]HandlerFunc // Map of all registered command handlers.
	states   map[int64]string       // Map of all active user states (active commands).
}

func New(client *tgclient.Client) *Manager {
	m := &Manager{
		client:   client,
		commands: make(map[string]HandlerFunc),
		states:   make(map[int64]string),
	}
	return m
}

func (m *Manager) RegisterCommand(cmd Handler) {
	prefix := "/" + cmd.Prefix()
	handler := cmd.ExecuteTelegram
	m.commands[prefix] = handler
	log.Printf("[INFO] %s registered \n", prefix)
}

func (m *Manager) Start() {
	cfg := tgclient.GetUpdatesConfig{
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

func (m *Manager) processMessage(message *tgclient.Message) {
	safeGet := func(arr []string, i int) string {
		if len(arr)-1 >= i {
			return arr[i]
		}
		return ""
	}

	s := strings.SplitN(message.Text, " ", 2)
	prefix := safeGet(s, 0)
	// subcommand := strings.Split(prefix, ":")
	args := safeGet(s, 1)

	fn := m.commands[prefix]
	if fn == nil {
		return
	}

	input := Input{}
	if message.From != nil {
		input.User = *message.From
	}
	input.Text = args

	res, err := fn(input)
	if err != nil {
		return
	}

	m.client.SendMessage(context.Background(), tgclient.SendMessageConfig{
		ChatId: message.Chat.ID,
		Text:   res.Text,
	})
}

func (m *Manager) processCallbackQuery(cq *tgclient.CallbackQuery) {

}
