package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
	tgclient "mr-weasel/client/telegram"
	"mr-weasel/commands/car"
	"mr-weasel/commands/ping"
	tgmanager "mr-weasel/manager/telegram"
	"os"
)

func main() {
	db := sqlx.MustConnect(os.Getenv("GOOSE_DRIVER"), os.Getenv("GOOSE_DBSTRING"))
	tgClient := tgclient.MustConnect(os.Getenv("TG_TOKEN"), false)
	tgManager := tgmanager.New(tgClient, db, true)
	tgManager.AddCommands(ping.PingCommand{}, car.CarCommand{})
	tgManager.PublishCommands()
	tgManager.Start()
}
