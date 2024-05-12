package handlers

import (
	"net/http"

	"github.com/vladeemerr/vibecheck/vick-ui/internal/components"
	"github.com/vladeemerr/vibecheck/vick-ui/internal/db"
)

type SignInGetHandler struct {
	AccessEmail string
}

type SignInPostHandler struct {
	Users *db.SimpleDB
}

func (h *SignInGetHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	c := components.SignIn(h.AccessEmail)
	err := c.Render(r.Context(), w)
	return err
}

func (h *SignInPostHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	r.ParseForm()

	login := r.FormValue("login")
	password := r.FormValue("password")

	key, exist := h.Users.Search(login)

	if !exist || key != password {
		w.WriteHeader(http.StatusUnauthorized)
		c := components.SignInFailed()
		return c.Render(r.Context(), w)
	}

	w.Header().Set("HX-Redirect", "/")

	return nil
}
