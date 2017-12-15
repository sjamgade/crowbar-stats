package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"crowbar-stats/config"
	"crowbar-stats/handler"
	"crowbar-stats/storage/sqlite"
)

func main() {

	configPath := flag.String("config", "./config/config.json", "path of the config file")

	flag.Parse()

	// Read config
	config, err := config.FromFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile(config.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	// Set use storage, select [Postgres, Filesystem, Redis ...]
	svc, err := sqlite.New(config.Sqlite.Dbpath)
	if err != nil {
		log.Fatal(err)
	}
	defer svc.Close()

	// Create a server
	http.Handle("/", handler.New(svc))
	server := &http.Server{Addr: fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port)}

	// Check for a closing signal
	go func() {
		// Graceful shutdown
		sigquit := make(chan os.Signal, 1)
		signal.Notify(sigquit, os.Interrupt, os.Kill)

		sig := <-sigquit
		log.Printf("caught sig: %+v", sig)
		log.Printf("Gracefully shutting down server...")

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("Unable to shut down server: %v", err)
		} else {
			log.Println("Server stopped")
		}
	}()

	// Start server
	log.Printf("Starting HTTP Server. Listening at %q", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("%v", err)
	} else {
		log.Println("Server closed!")
	}
}
