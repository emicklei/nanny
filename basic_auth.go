package nanny

import "net/http"

// BasicAuthHandler is a http.Handler that requires basic authentication.
// 95% Suggested by Google Duet
type BasicAuthHandler struct {
	Handler  http.Handler
	Username string
	Password string
}

func NewBasicAuthHandler(handler http.Handler, username, password string) *BasicAuthHandler {
	return &BasicAuthHandler{
		Handler:  handler,
		Username: username,
		Password: password,
	}
}

func (h *BasicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok || user != h.Username || pass != h.Password {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}
	h.Handler.ServeHTTP(w, r)
}
