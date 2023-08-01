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
	_ = sqlx.MustConnect("sqlite3", "bin/mr-weasel.db")
	// db.MustExec("INSERT INTO car (user_id, name, year, plate) values (?, ?, ?, ?)", 1, "BMW", 2021, "FZ-28")
	token := os.Getenv("TG_TOKEN")
	tgClient := tgclient.MustConnect(token, false)
	tgManager := tgmanager.New(tgClient, true)
	tgManager.AddCommands(ping.PingCommand{}, car.CarCommand{})
	tgManager.PublishCommands()
	tgManager.Start()
}
