package index

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
)

func NewHandler(sm *scs.SessionManager) http.Handler {
	return &handler{}
}

type handler struct {
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	component := page(r)
	component.Render(r.Context(), w)
}
