package handlers

import (
	"net/http"

	"github.com/vladeemerr/vibecheck/vick-ui/internal/components"
)

type DashboardHandler struct {}

func (h *DashboardHandler) HandleGet(w http.ResponseWriter, r *http.Request) error {
	c := components.Base("Dashboard")
	err := c.Render(r.Context(), w)
	return err
}
