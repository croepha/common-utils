package logging

import (
	"log/slog"
	"os"
)

// Perform some common startup things
func SlogStartup() {
	// TODO: It would be nicer if we could connect slog to t.Log when running tests...

	// Enable debug logs if the env is set:
	if os.Getenv("SLOG_DEBUG") == "1" {
		slog.SetDefault(slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
	}

}
