package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
)

var (
	exitCallbacks = make(map[string]func())
)

func cleanup() {
	for _, cb := range exitCallbacks {
		cb()
	}
}

func main() {
	flag.Parse()

	defer cleanup()

	doneCh := make(chan os.Signal, 1)
	signal.Notify(doneCh, os.Interrupt)
	go func() {
		<-doneCh
		cleanup()
	}()

	switch flag.Arg(0) {
	case "ssh":
		log.Fatal(serveSSH())
	case "http":
		log.Fatal(serveHTTP())
	default:
		os.Exit(1)
		return
	}

}
