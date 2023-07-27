package main

import (
	_ "github.com/joho/godotenv/autoload"
	tgclient "mr-weasel/client/telegram"
	tgmanager "mr-weasel/manager/telegram"
	"os"
)

import (
	"mr-weasel/commands/car"
	"mr-weasel/commands/ping"
)

func main() {
	token := os.Getenv("TG_TOKEN")
	tgClient, err := tgclient.New(token, false)
	if err != nil {
		panic(err)
	}
	tgManager := tgmanager.New(tgClient, true)
	tgManager.AddCommands(ping.PingCommand{}, car.CarCommand{})
	tgManager.PublishCommands()
	tgManager.Start()
}
