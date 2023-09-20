package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"

	"mr-weasel/commands"
	"mr-weasel/storage"
	"mr-weasel/telegram"
	"mr-weasel/utils"

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

	queuePool, err1 := strconv.Atoi(os.Getenv("QUEUE_POOL"))
	queueParallel, err2 := strconv.Atoi(os.Getenv("QUEUE_PARALLEL"))
	if errors.Join(err1, err2) != nil {
		panic("Invalid QUEUE_POOL or QUEUE_PARALLEL values")
	}

	queue := utils.NewQueue(queuePool, queueParallel)

	tgClient := telegram.MustConnect(os.Getenv("TG_TOKEN"), false)
	tgManager := telegram.NewManager(tgClient)

	if os.Getenv("RTX_MODE") == "on" {
		commands := tgManager.AddCommands(
			commands.NewExtractVoiceCommand(queue),
		)
		tgManager.PublishCommands(commands)
	} else {
		commands := tgManager.AddCommands(
			commands.NewPingCommand(),
			commands.NewCarCommand(storage.NewCarStorage(db)),
			commands.NewHolidayCommand(storage.NewHolidayStorage(db)),
			commands.NewExtractVoiceCommand(queue),
		)
		tgManager.PublishCommands(commands)
	}

	tgManager.Start(mainContext())
}

func mainContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx
}
