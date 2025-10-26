package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nebojsaj1726/proxy-pool/app"
)

func main() {
	app, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	go func() {
		if err := app.Start(); err != nil {
			log.Fatalf("App error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	app.Stop()
}
