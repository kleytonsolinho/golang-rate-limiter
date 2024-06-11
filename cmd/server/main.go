package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/kleytonsolinho/golang-rate-limiter/configs"
	"github.com/kleytonsolinho/golang-rate-limiter/internal/infra/database"
	"github.com/kleytonsolinho/golang-rate-limiter/utils"
)

func main() {
	config, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	rdb := database.LoadStorage(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	rateLimitPerSecond, _ := strconv.Atoi(config.RateLimitPerSecond)
	rateLimitWithTokenPerSecond, _ := strconv.Atoi(config.RateLimitWithTokenPerSecond)
	rateLimitBlockDurationInMinutes, _ := strconv.Atoi(config.RateLimitBlockDurationInMinutes)
	blockDuration := time.Duration(rateLimitBlockDurationInMinutes) * time.Minute // Tempo de bloqueio após atingir o limite

	limitByIP := httprate.Limit(
		rateLimitPerSecond, // requesições/seg por IP
		1*time.Second,      // por duração
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			ip := utils.GetIP(r)
			err := rdb.Create(ip, "1", blockDuration)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
		}),
	)

	limitByToken := httprate.Limit(
		rateLimitWithTokenPerSecond, // requesições/seg por Token
		1*time.Second,               // por duração
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Ratelimit-Token")
			err := rdb.Create(token, "1", blockDuration)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
		}),
		httprate.WithKeyFuncs(
			func(r *http.Request) (string, error) {
				token := r.Header.Get("X-Ratelimit-Token")
				if token != "" {
					return token, nil
				}
				return "invalid", nil
			},
		),
	)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := utils.GetIP(r)
			token := r.Header.Get("X-Ratelimit-Token")

			// Verificar se o IP ou Token está bloqueado
			blocked, err := rdb.Exists(ip, token)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if blocked {
				http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
				return
			}

			// Aplica o rate limit adequado
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
