package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"mr-weasel/internal/bot"
	"mr-weasel/internal/commands"
	"mr-weasel/internal/config"
	"mr-weasel/internal/lib/db"
	"mr-weasel/internal/lib/logger"
	"mr-weasel/internal/lib/queue"
	"mr-weasel/internal/lib/telegram"
	"mr-weasel/internal/storage"
	"mr-weasel/internal/utils"

	"mr-weasel/migrations"
)

func main() {
	config := config.GetConfig()

	if config.Debug {
		logger.SetLevel(slog.LevelDebug)
	}

	store, err := db.NewStore(config.DBDriver, config.DBString)
	if err != nil {
		logger.GetLogger().Error("unable to connect to the database", "err", err)
		os.Exit(1)
	}

	if err := db.MigrateUp(logger.GetLogger(), config.DBDriver, store.DB(), migrations.MigrationsFS, "."); err != nil {
		logger.GetLogger().Error("unable to migrate the database", "err", err)
		os.Exit(1)
	}

	queue := queue.NewQueue(config.QueuePool, config.QueueParallel)

	audioSeparator := utils.NewAudioSeparator()
	voiceChanger := utils.NewVoiceChanger()

	tgClient, err := telegram.Connect(config.TGToken)
	if err != nil {
		logger.GetLogger().Error("unable to connect to the telegram", "err", err)
		os.Exit(1)
	}

	botManager := bot.NewManager(tgClient)

	if config.RTXMode {
		commands := botManager.AddCommands(
			commands.NewPingCommand(),
			commands.NewYTMP3Command(),
			commands.NewExtractVoiceCommand(queue, audioSeparator),
			commands.NewChangeVoiceCommand(storage.NewRvcStorage(store.DBX()), queue, audioSeparator, voiceChanger),
		)
		botManager.PublishCommands(commands)
	} else {
		commands := botManager.AddCommands(
			commands.NewPingCommand(),
			commands.NewCarCommand(storage.NewCarStorage(store.DBX())),
			commands.NewHolidayCommand(storage.NewHolidayStorage(store.DBX())),
			commands.NewYTMP3Command(),
		)
		botManager.PublishCommands(commands)
	}

	ctx := mainContext()
	botManager.Start(ctx)
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
