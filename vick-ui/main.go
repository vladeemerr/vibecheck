package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"context"
	"errors"
)

func main() {
	exit := make(chan os.Signal)
	signal.Notify(exit, os.Interrupt, os.Kill)

	mux := http.NewServeMux()

	server := http.Server{
		Addr: ":3000",
		Handler: mux,
	}

	log.Println("Attempting to start a server")

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Println("Server was closed")
			} else {
				log.Fatalln("Server error occurred:", err)
			}
		}
	}()

	log.Println("Server started on", server.Addr)

	<-exit

	log.Println("Attempting to shut down server")

	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalln("Server failed to shut down:", err)
	}
}
