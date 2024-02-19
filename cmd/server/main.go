package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGTERM|syscall.SIGINT|syscall.SIGHUP)
	app, err := NewChatServer()
	if err != nil {
		log.Fatal(err)
	}

	app.GoServe()

	// wait for exit signal
	<-exitSignal

	// stop server
	app.Stop()

	fmt.Println("application gracefully terminated")
}
