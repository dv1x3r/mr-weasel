package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
	"mr-weasel/commands"
	"mr-weasel/storage"
	"mr-weasel/telegram"
	"os"
)

func main() {
	db := sqlx.MustConnect(os.Getenv("GOOSE_DRIVER"), os.Getenv("GOOSE_DBSTRING"))
	tgClient := telegram.MustConnect(os.Getenv("TG_TOKEN"), false)
	tgManager := telegram.NewManager(tgClient)
	tgManager.AddCommands(
		commands.NewPingCommand(),
		commands.NewCarCommand(storage.NewCarStorage(db)),
	)
	tgManager.PublishCommands()
	tgManager.Start()
}
