package main

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"mr-weasel/api/telegram"
	"os"
	"time"
)

func main() {
	token := os.Getenv("TG_TOKEN")
	tg, err := telegram.New(token, 120*time.Second)
	if err != nil {
		panic(err)
	}
	log.Printf("%+v \n", tg.Me)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := telegram.GetUpdatesConfig{Timeout: 60}
	updates := tg.GetUpdatesChan(ctx, cfg, 100)
	for update := range updates {
		log.Println(update)
	}

}
