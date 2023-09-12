package main

import (
	"os"

	"mr-weasel/commands"
	"mr-weasel/storage"
	"mr-weasel/telegram"

	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbDriver, dbString := os.Getenv("GOOSE_DRIVER"), os.Getenv("GOOSE_DBSTRING")
	if dbDriver == "sqlite3" {
		dbString += "?_journal=WAL&_fk=1"
	}
	db := sqlx.MustConnect(dbDriver, dbString)
	tgClient := telegram.NewClient(os.Getenv("TG_TOKEN"), false).MustConnect()
	tgManager := telegram.NewManager(tgClient)
	if os.Getenv("RTX_MODE") == "on" {
		tgManager.AddCommands(
			commands.NewPythonCommand(),
		)
	} else {
		tgManager.AddCommands(
			commands.NewPingCommand(),
			commands.NewCarCommand(storage.NewCarStorage(db)),
			commands.NewHolidayCommand(storage.NewHolidayStorage(db)),
		)
	}
	tgManager.PublishCommands()
	tgManager.Start()
}
