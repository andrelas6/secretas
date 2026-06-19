package health

import "net/http"

type HealthzHandler struct {
	// when this needs deps, it will come here
}

var _ http.Handler = HealthzHandler{}

func (h HealthzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
