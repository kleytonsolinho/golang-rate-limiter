package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/kleytonsolinho/golang-rate-limiter/configs"
)

func main() {
	config, _ := configs.LoadConfig(".")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	rateLimitPerSecond, _ := strconv.Atoi(config.RateLimitPerSecond)

	r.Use(httprate.Limit(
		rateLimitPerSecond, // requests
		1*time.Second,      // per duration
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
		}),
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			return r.Header.Get("X-Access-Token"), nil
		}),
	))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Success!"))
	})

	http.ListenAndServe(config.WebServerPort, r)
}
