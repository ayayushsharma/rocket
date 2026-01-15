package main

import (
	"log/slog"
	"os"

	"ayayushsharma/rocket/cmd"
)

func main() {
	var programLevel slog.LevelVar
	programLevel.Set(slog.LevelDebug)

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: &programLevel,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	cmd.Execute()
}
