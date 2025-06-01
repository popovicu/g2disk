package main

import (
	"flag"
	"log/slog"
)

func main() {
	flag.Parse()
	logger := slog.Default()

	logger.Info("Hello world!")
}
