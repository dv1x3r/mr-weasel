package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

func splitCommand(input string, prefix string) []string {
	input, _ = strings.CutPrefix(input, prefix)
	input = strings.TrimSpace(input)
	if input == "" {
		return []string{}
	}
	return strings.Split(input, " ")
}

func commandf(h Handler, args ...any) string {
	cmd := h.Prefix()
	for _, arg := range args {
		cmd = fmt.Sprintf("%s %v", cmd, arg)
	}
	return cmd
}

func cancelf(ctx context.Context) string {
	return fmt.Sprintf("%s %s", CmdCancel, ctx.Value("contextID"))
}

func safeGet(args []string, n int) string {
	if n <= len(args)-1 {
		return args[n]
	}
	return ""
}

func safeGetInt(args []string, n int) int {
	i, _ := strconv.Atoi(safeGet(args, n))
	return i
}

func safeGetInt64(args []string, n int) int64 {
	return int64(safeGetInt(args, n))
}
