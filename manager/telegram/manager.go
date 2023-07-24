package tgmanager

import (
	"context"
	"log"
	"mr-weasel/client/telegram"
	"strings"
)

type Input struct {
	Prefix string
	Action string
	Args   string
	User   tgclient.User
}

type Output struct {
	Text  string
	State string
}

type Handler interface {
	Prefix() string
	Description() string
	ExecuteTelegram(Input) (Output, error)
}

type HandlerFunc = func(Input) (Output, error)

type Manager struct {
	client   *tgclient.Client       // Telegram API Client
	debug    bool                   // Enable debug output.
	commands map[string]HandlerFunc // Map of all registered command handlers.
	states   map[int64]string       // Map of all active user states (active commands).
}

func New(client *tgclient.Client, debug bool) *Manager {
	m := &Manager{
		client:   client,
		debug:    debug,
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

// Parse Telegram message in the following format: /prefix:action args
func parseMessage(message *tgclient.Message) Input {
	safeGet := func(arr []string, i int) string {
		if len(arr)-1 >= i {
			return arr[i]
		}
		return ""
	}

	textSplit := strings.SplitN(message.Text, " ", 2)
	command, args := safeGet(textSplit, 0), safeGet(textSplit, 1)

	commandSplit := strings.Split(command, ":")
	prefix, action := safeGet(commandSplit, 0), safeGet(commandSplit, 1)

	var user tgclient.User
	if message.From != nil {
		user = *message.From
	}

	return Input{Prefix: prefix, Action: action, Args: args, User: user}
}

func (m *Manager) processMessage(message *tgclient.Message) {
	input := parseMessage(message)

	fn := m.commands[input.Prefix]
	if fn == nil {
		if m.debug {
			log.Println("[DEBUG]", input.Prefix, "handler not found")
		}
		return
	}

	res, err := fn(input)
	if err != nil {
		log.Println("[ERROR]", input.Prefix, err)
		return
	}

	m.client.SendMessage(context.Background(), tgclient.SendMessageConfig{
		ChatId: message.Chat.ID,
		Text:   res.Text,
	})

	if m.debug {
		log.Println("[DEBUG]", input.Prefix, "succeeded")
	}
}

func (m *Manager) processCallbackQuery(cq *tgclient.CallbackQuery) {

}
