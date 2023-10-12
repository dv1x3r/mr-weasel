package tgmanager

import (
	"context"
	"testing"

	"mr-weasel/commands"
)

type TestCommand struct {
}

func (TestCommand) Prefix() string {
	return "/test"
}

func (TestCommand) Description() string {
	return "Test Command"
}

func (tc *TestCommand) Execute(ctx context.Context, pl commands.Payload) {
	if pl.Command == "state" {
		pl.ResultChan <- commands.Result{State: tc.ExecuteState}
	}
	pl.ResultChan <- commands.Result{}
}

func (tc *TestCommand) ExecuteState(ctx context.Context, pl commands.Payload) {
	pl.ResultChan <- commands.Result{}
}

func TestGetHandlerFunc(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		manager := NewManager(nil)
		fn, ok := manager.getExecuteFunc(0, "/test")
		if fn != nil || ok {
			t.Fatalf("expected [nil, false], actual [%p %t]\n", fn, ok)
		}
	})

	t.Run("Found", func(t *testing.T) {
		manager := NewManager(nil)
		manager.AddCommands(new(TestCommand))
		fn, ok := manager.getExecuteFunc(0, "/test")
		if fn == nil || !ok {
			t.Fatalf("expected [nil, false], actual [%p %t]\n", fn, ok)
		}
	})
}
