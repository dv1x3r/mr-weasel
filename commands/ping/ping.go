package ping

import (
	"context"
	"github.com/jmoiron/sqlx"
	tg "mr-weasel/manager/telegram"
)

type PingCommand struct{}

func (PingCommand) Prefix() string {
	return "ping"
}

func (PingCommand) Description() string {
	return "answer with pong"
}

func (PingCommand) Execute(ctx context.Context, db *sqlx.DB, pl tg.Payload) (tg.Result, error) {
	if pl.Command == "/ping me" {
		return tg.Result{Text: "What is your name?", State: personalized}, nil
	}
	return tg.Result{Text: "pong!"}, nil
}

func personalized(ctx context.Context, db *sqlx.DB, pl tg.Payload) (tg.Result, error) {
	return tg.Result{Text: "Pong to " + pl.Command + "!"}, nil
}
