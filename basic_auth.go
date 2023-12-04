package nanny

import (
	"log/slog"
	"net/http"
)

// BasicAuthHandler is a http.Handler that requires basic authentication.
// 95% Suggested by Google Duet
type BasicAuthHandler struct {
	Handler  http.Handler
	Username string
	Password string
}

func NewBasicAuthHandler(handler http.Handler, username, password string) *BasicAuthHandler {
	// is it configured correctly?
	if username == "" || password == "" {
		slog.Warn("nanny.BasicAuthHandler is not configured correctly (missing username or password)")
	}
	return &BasicAuthHandler{
		Handler:  handler,
		Username: username,
		Password: password,
	}
}

func (h *BasicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// was it configured correctly?
	if h.Username == "" || h.Password == "" {
		h.Handler.ServeHTTP(w, r)
		return
	}
	user, pass, ok := r.BasicAuth()
	if !ok || user != h.Username || pass != h.Password {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}
	h.Handler.ServeHTTP(w, r)
}
