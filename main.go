package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/maanijou/imageservice/database"
	"github.com/maanijou/imageservice/image"
	"github.com/maanijou/imageservice/middleware"
	"github.com/maanijou/imageservice/monitoring"
)

func main() {
	log.Println("Running Image app.")
	err := database.InitDatabase("image")
	if err != nil {
		log.Fatal("Cannot initiate the database. exiting...")
	}
	err = database.GlobalDB.AutoMigrate(&image.Image{})
	if err != nil {
		log.Fatal("Cannot migrate image database. exiting...")
	}
	sm := mux.NewRouter().StrictSlash(true) // ignoring trailing slash
	sm = sm.PathPrefix("/api/v1/").Subrouter()
	sm.Use(middleware.LoggingMiddleware)
	sm.Use(middleware.CorsMiddleware)

	monitoring.SetupMonitoring(sm)
	image.SetupRoutes(sm)

	s := http.Server{
		Addr:         ":8080",           // configure the bind address
		Handler:      sm,                // set the default handler
		ReadTimeout:  5 * time.Second,   // max time to read request from the client
		WriteTimeout: 10 * time.Second,  // max time to write response to the client
		IdleTimeout:  120 * time.Second, // max time for connections using TCP Keep-Alive
	}

	go func() {
		log.Println("Starting server on port 8080")

		err := s.ListenAndServe()
		if err != nil {
			log.Printf("Error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	// trap sigterm or interupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	// Block until a signal is received.
	sig := <-c
	log.Println("Got signal:", sig)

	// gracefully shutdown the server, waiting max 30 seconds for current operations to complete
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(ctx)
}
