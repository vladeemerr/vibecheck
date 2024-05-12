package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/vladeemerr/vibecheck/vick-ui/internal/components"
	"github.com/vladeemerr/vibecheck/vick-ui/internal/storage"
)

type SignInGetHandler struct {
	AccessEmail string
}

type SignInPostHandler struct {
	Users *storage.SimpleStorage
	Sessions *storage.SimpleStorage
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

	sessionBytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, sessionBytes); err != nil {
		return err
	}

	session := base64.URLEncoding.EncodeToString(sessionBytes)
	h.Sessions.Insert(login, session)

	w.Header().Set("HX-Redirect", "/")

	http.SetCookie(w, &http.Cookie{
		Name: "session",
		Value: session,
		Path: "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}
