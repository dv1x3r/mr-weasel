package main

import (
	_ "github.com/joho/godotenv/autoload"
	tgclient "mr-weasel/client/telegram"
	tgmanager "mr-weasel/manager/telegram"
	"os"
)

func main() {
	token := os.Getenv("TG_TOKEN")
	tgClient, err := tgclient.New(token)
	if err != nil {
		panic(err)
	}
	tgManager := tgmanager.New(tgClient)
	tgManager.Start()
}
