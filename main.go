package main

import (
	// "github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	// _ "github.com/mattn/go-sqlite3"
	// "log"
	tgclient "mr-weasel/client/telegram"
	"mr-weasel/commands/car"
	"mr-weasel/commands/ping"
	tgmanager "mr-weasel/manager/telegram"
	"os"
)

func main() {
	// db, err := sqlx.Open("sqlite3", "bin/mr-weasel.db")
	// if err != nil {
	// 	panic(err)
	// }
	// err = db.Ping()
	// log.Println("DB ERR", err)
	// db.MustExec("INSERT INTO car (user_id, name, year, plate) values (?, ?, ?, ?)", 1, "BMW", 2021, "FZ-28")

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
