package middleware_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/maanijou/imageservice/middleware"
	"github.com/maanijou/imageservice/monitoring"
)

var sm *mux.Router

func TestMain(m *testing.M) {
	route, err := setup()
	sm = route
	if err != nil {
		teardown()
		os.Exit(-1)
		return
	}
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() (*mux.Router, error) {
	log.Println("Setup middleware test")
	os.Chdir("..")
	sm := mux.NewRouter().StrictSlash(true) // ignoring trailing slash
	sm = sm.PathPrefix("/api/v1/").Subrouter()
	sm.Use(middleware.CorsMiddleware)
	sm.Use(middleware.LoggingMiddleware)
	monitoring.SetupMonitoring(sm)

	return sm, nil
}

func teardown() {
	log.Println("cleanup middleware test")
}

func TestCORS(t *testing.T) {
	req, err := http.NewRequest(
		"GET",
		"/api/v1/health/",
		nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.Handler(sm)
	handler.ServeHTTP(rr, req)

	if rr.HeaderMap.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf(
			"Error in allow origin expected '*', got %v",
			rr.HeaderMap.Get("Access-Control-Allow-Origin"),
		)
	}
}
