package tests

import (
	"log/slog"
	"os"
)

// fileLogger is a slog.Logger that writes to a file.
type fileLogger struct {
	file *os.File
	*slog.Logger
}

func newFileLogger(path string) (fileLogger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	l := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelError}))
	return fileLogger{file: f, Logger: l}, nil
}

// Close closes the underlying file.
func (f fileLogger) Close() error {
	return f.file.Close()
}
