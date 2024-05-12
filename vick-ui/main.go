package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"context"
	"errors"

	"github.com/vladeemerr/vibecheck/vick-ui/internal/handlers"
	"github.com/vladeemerr/vibecheck/vick-ui/internal/storage"
)

func main() {
	exit := make(chan os.Signal)
	signal.Notify(exit, os.Interrupt, os.Kill)

	mux := http.NewServeMux()

	users := storage.NewSimpleStorage()
	users.Insert("example", "1234")

	sessions := storage.NewSimpleStorage()

	fs := http.FileServer(http.Dir("./vick-ui/assets/"))
	mux.Handle("GET /assets/*", http.StripPrefix("/assets/", fs))

	dashboard := handlers.DashboardHandler{}
	mux.HandleFunc("GET /", func (w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Sorry, wrong room"))
			return
		}

		handlers.Handler(dashboard.HandleGet).ServeHTTP(w, r)
	})

	signinGet := handlers.SignInGetHandler{
		AccessEmail: "localhost@localdomain",
	}
	mux.Handle("GET /signin", handlers.Handler(signinGet.Handle))

	signinPost := handlers.SignInPostHandler{
		Users: users,
		Sessions: sessions,
	}
	mux.Handle("POST /signin", handlers.Handler(signinPost.Handle))

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
