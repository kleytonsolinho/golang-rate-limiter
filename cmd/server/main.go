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
	config, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	rateLimitPerSecond, _ := strconv.Atoi(config.RateLimitPerSecond)
	rateLimitWithTokenPerSecond, _ := strconv.Atoi(config.RateLimitWithTokenPerSecond)

	limitByIP := httprate.Limit(
		rateLimitPerSecond, // requests per IP
		1*time.Second,      // per duration
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
		}),
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			token := r.Header.Get("X-Ratelimit-Token")
			if token == "" {
				return r.RemoteAddr, nil
			}
			return "", nil
		}),
	)

	limitByToken := httprate.Limit(
		rateLimitWithTokenPerSecond, // requests per token
		1*time.Second,               // per duration
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
		}),
		httprate.WithKeyFuncs(
			func(r *http.Request) (string, error) {
				token := r.Header.Get("X-Ratelimit-Token")
				if token != "" {
					return token, nil
				}
				return "", nil
			},
		),
	)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Ratelimit-Token")

			if token == "" {
				limitByIP(next).ServeHTTP(w, r)
			} else {
				limitByToken(next).ServeHTTP(w, r)
			}
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Success!"))
	})

	http.ListenAndServe(":"+config.WebServerPort, r)
}
