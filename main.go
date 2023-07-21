package main

import (
	_ "github.com/joho/godotenv/autoload"
	"mr-weasel/api/telegram"
	"mr-weasel/manager"
	"os"
)

func main() {
	token := os.Getenv("TG_TOKEN")
	tg, err := telegram.New(token)
	if err != nil {
		panic(err)
	}
	tgmng := manager.NewTelegramManager(tg)
	tgmng.Start()
}
