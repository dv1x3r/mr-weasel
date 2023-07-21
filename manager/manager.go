package manager

import (
	"context"
	"log"
	"mr-weasel/api/telegram"
)

type TelegramManager struct {
	client         *telegram.Client
	activeCommands map[string]string
}

func NewTelegramManager(client *telegram.Client) *TelegramManager {
	return &TelegramManager{client: client}
}

func (mng *TelegramManager) Start() {
	cfg := telegram.GetUpdatesConfig{Timeout: 60}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	updates := mng.client.GetUpdatesChan(ctx, cfg, 100)
	for update := range updates {
		log.Println(update)
	}
}
