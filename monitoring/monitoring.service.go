package monitoring

import (
	"net/http"

	"github.com/gorilla/mux"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

func SetupMonitoring(sm *mux.Router) {
	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/health", healthCheck)
}
