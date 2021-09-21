package model

import (
	"os"
	"os/signal"
	"syscall"
)

func hookSignals() chan os.Signal {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, syscall.SIGTERM)
	return sigchan
}
