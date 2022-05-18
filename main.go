package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/jjlock/bank/server"
)

func main() {
	s := server.New()
	dbFile := "db.gob"
	port := "8000"

	log.Println("Loading data from database file if present")
	if err := s.LoadDB(dbFile); err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("Failed to load database from %s: %v", dbFile, err)
		}
	}

	srv := &http.Server{Addr: ":" + port}

	// routes
	http.HandleFunc("/", s.LoginHandler)
	http.HandleFunc("/login", s.LoginHandler)
	http.HandleFunc("/signup", s.SignupHandler)
	http.HandleFunc("/account", s.AccountHandler)
	http.HandleFunc("/transaction", s.TransactionHandler)
	http.HandleFunc("/logout", s.LogoutHandler)

	// static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	idleConnsClosed := make(chan struct{})
	go func() {
		// listen for ctrl+C
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		log.Println("Server shutting down")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Println("Starting server at http://localhost:" + port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	log.Println("Waiting for idle connections to close")
	<-idleConnsClosed

	log.Println("Saving database to file")
	if err := s.SaveDB("db.gob"); err != nil {
		log.Printf("Failed to save data to database: %v", err)
	}

	log.Println("Server shutdown")
}
