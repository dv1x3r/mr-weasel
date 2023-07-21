package tgmanager

import (
	"context"
	"log"
	"mr-weasel/client/telegram"
)

type Manager struct {
	client         *tgclient.Client
	activeCommands map[string]string
}

func New(client *tgclient.Client) *Manager {
	return &Manager{client: client}
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
		log.Println(update)
		if update.Message != nil {
			m.processMessage(update.Message)
		}
		if update.CallbackQuery != nil {
			m.processCallbackQuery(update.CallbackQuery)
		}
	}
}

func (m *Manager) processMessage(message *tgclient.Message) {

}

func (m *Manager) processCallbackQuery(cq *tgclient.CallbackQuery) {

}
