package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"

	"mr-weasel/commands"
	"mr-weasel/storage"
	"mr-weasel/tgclient"
	"mr-weasel/tgmanager"
	"mr-weasel/utils"

	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

func main() {
	dbDriver, dbString := os.Getenv("DB_DRIVER"), os.Getenv("DB_STRING")
	var db *sqlx.DB

	if dbDriver == "sqlite3" {
		// mattn/go-sqlite3 (cgo)
		db = sqlx.MustConnect(dbDriver, dbString+"?_journal=WAL&_fk=1&_busy_timeout=10000")
	} else if dbDriver == "sqlite" {
		// modernc.org/sqlite (cgo-free)
		// db = sqlx.MustConnect(dbDriver, dbString+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(10000)")
		panic(errors.New("modernc.org/sqlite (cgo-free) is not supported, switch to mattn/go-sqlite3"))
	} else {
		panic(errors.New(dbDriver + " (db driver) is not supported, switch to mattn/go-sqlite3"))
	}

	if err := goose.SetDialect(dbDriver); err != nil {
		panic(err)
	}

	if err := goose.Up(db.DB, "./migrations"); err != nil {
		panic(err)
	}

	queuePool, err1 := strconv.Atoi(os.Getenv("QUEUE_POOL"))
	queueParallel, err2 := strconv.Atoi(os.Getenv("QUEUE_PARALLEL"))
	if errors.Join(err1, err2) != nil {
		panic("Invalid QUEUE_POOL or QUEUE_PARALLEL values")
	}

	queue := utils.NewQueue(queuePool, queueParallel)
	audioSeparator := utils.NewAudioSeparator()
	voiceChanger := utils.NewVoiceChanger()

	tgClient := tgclient.MustConnect(os.Getenv("TG_TOKEN"), false)
	tgManager := tgmanager.NewManager(tgClient)

	if os.Getenv("RTX_MODE") == "on" {
		commands := tgManager.AddCommands(
			commands.NewPingCommand(),
			commands.NewYTMP3Command(),
			commands.NewExtractVoiceCommand(queue, audioSeparator),
			commands.NewChangeVoiceCommand(storage.NewRvcStorage(db), queue, audioSeparator, voiceChanger),
		)
		tgManager.PublishCommands(commands)
	} else {
		commands := tgManager.AddCommands(
			commands.NewPingCommand(),
			commands.NewCarCommand(storage.NewCarStorage(db)),
			commands.NewHolidayCommand(storage.NewHolidayStorage(db)),
			commands.NewYTMP3Command(),
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
