package handlers

import (
	"log"
	"net/http"
)

type HandlerFunc func (w http.ResponseWriter, r *http.Request) error

func Handler(h HandlerFunc) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			log.Println("Handler error occurred:", err)
		}
	}
}
