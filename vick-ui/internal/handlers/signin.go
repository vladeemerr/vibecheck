package handlers

import (
	"log"
	"net/http"

	"github.com/vladeemerr/vibecheck/vick-ui/internal/components"
)

type SignInHandler struct {
	AccessEmail string
}

func (h *SignInHandler) HandleGet(w http.ResponseWriter, r *http.Request) error {
	c := components.SignIn(h.AccessEmail)
	err := c.Render(r.Context(), w)
	return err
}

func (h *SignInHandler) HandlePost(w http.ResponseWriter, r *http.Request) error {
	r.ParseForm()

	login := r.FormValue("login")
	password := r.FormValue("password")

	log.Println(login, password)

	w.WriteHeader(http.StatusUnauthorized)
	c := components.SignInFailed()
	err := c.Render(r.Context(), w)

	return err
}
