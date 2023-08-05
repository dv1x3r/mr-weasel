package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
	"mr-weasel/commands"
	"mr-weasel/storage"
	"mr-weasel/tgclient"
	"mr-weasel/tgmanager"
	"os"
)

func main() {
	db := sqlx.MustConnect(os.Getenv("GOOSE_DRIVER"), os.Getenv("GOOSE_DBSTRING"))
	tgClient := tgclient.MustConnect(os.Getenv("TG_TOKEN"), true)
	tgManager := tgmanager.New(tgClient)
	tgManager.AddCommands(
		commands.NewPingCommand(),
		commands.NewCarCommand(storage.NewCarStorage(db)),
	)
	tgManager.PublishCommands()
	tgManager.Start()
}
