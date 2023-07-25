package main

import (
	tgclient "mr-weasel/client/telegram"
	"mr-weasel/commands/car"
	"mr-weasel/commands/ping"
	tgmanager "mr-weasel/manager/telegram"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	token := os.Getenv("TG_TOKEN")
	tgClient, err := tgclient.New(token, true)
	if err != nil {
		panic(err)
	}
	tgManager := tgmanager.New(tgClient, true)
	tgManager.AddCommand(ping.PingCommand{})
	tgManager.AddCommand(car.CarCommand{})
	tgManager.SetCommands()
	tgManager.Start()
}
