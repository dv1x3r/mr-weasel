package tgmanager

import (
	"context"
	"log"
	"mr-weasel/client/telegram"
	"strings"
)

type Command struct {
	UserID int64
	Prefix string
	Action string
	Args   string
}

type Result struct {
	Text   string
	Action string
}

type Handler interface {
	Prefix() string
	Description() string
	ExecuteTelegram(Command) (Result, error)
}

type HandlerFunc = func(Command) (Result, error)

type Manager struct {
	client   *tgclient.Client       // Telegram API Client
	debug    bool                   // Enable debug output.
	commands map[string]HandlerFunc // Map of all registered command handlers.
	states   map[int64]Command      // Map of all active user states (active commands).
}

func New(client *tgclient.Client, debug bool) *Manager {
	return &Manager{
		client:   client,
		debug:    debug,
		commands: make(map[string]HandlerFunc),
		states:   make(map[int64]Command),
	}
}

func (m *Manager) RegisterCommand(h Handler) {
	prefix := "/" + h.Prefix()
	m.commands[prefix] = h.ExecuteTelegram
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

func (m *Manager) updateState(cmd Command, res Result) {
	if res.Action != "" {
		cmd.Action = res.Action
		m.states[cmd.UserID] = cmd
	} else {
		delete(m.states, cmd.UserID)
	}
}

func readCommand(msg *tgclient.Message) Command {
	safeGet := func(arr []string, i int) string {
		if len(arr)-1 >= i {
			return arr[i]
		}
		return ""
	}

	s := strings.SplitN(msg.Text, " ", 2)
	command, args := safeGet(s, 0), safeGet(s, 1)

	s = strings.Split(command, ":")
	prefix, action := safeGet(s, 0), safeGet(s, 1)

	if msg.From == nil {
		return Command{}
	}

	return Command{UserID: msg.From.ID, Prefix: prefix, Action: action, Args: args}
}

func (m *Manager) getCommandHandler(msg *tgclient.Message) (Command, HandlerFunc) {
	cmd := readCommand(msg)          // Split message by /prefix:action args
	fn, ok := m.commands[cmd.Prefix] // Get the command handler (if exists)
	if ok {
		return cmd, fn
	}

	cmd, ok = m.states[msg.From.ID] // Check if user has an active state
	if ok {
		fn = m.commands[cmd.Prefix] // Get the command handler for that state
		cmd.Args = msg.Text         // Set text from the message input as args
		return cmd, fn
	}

	return Command{}, nil
}

func (m *Manager) processMessage(msg *tgclient.Message) {
	cmd, fn := m.getCommandHandler(msg)
	if fn == nil {
		if m.debug {
			log.Println("[DEBUG] Handler not found:", msg.Text)
		}
		return
	}

	res, err := fn(cmd)
	if err != nil {
		log.Printf("[ERROR] %+v %s \n", cmd, err)
		return
	}

	m.updateState(cmd, res) // Manage stateful commands
	log.Printf("[INFO] %+v succeeded \n", cmd)

	_, err = m.client.SendMessage(context.Background(), tgclient.SendMessageConfig{
		ChatId: msg.Chat.ID,
		Text:   res.Text,
	})
	if err != nil {
		log.Println("[ERROR] Sending a response:", err)
	}
}

func (m *Manager) processCallbackQuery(cq *tgclient.CallbackQuery) {

}
