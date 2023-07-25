package main

import (
	_ "github.com/joho/godotenv/autoload"
	tgclient "mr-weasel/client/telegram"
	"mr-weasel/commands/ping"
	tgmanager "mr-weasel/manager/telegram"
	"os"
)

func main() {
	token := os.Getenv("TG_TOKEN")
	tgClient, err := tgclient.New(token, true)
	if err != nil {
		panic(err)
	}
	tgManager := tgmanager.New(tgClient, true)
	tgManager.RegisterCommand(ping.PingCommand{})
	tgManager.Start()
}
