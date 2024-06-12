package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/kleytonsolinho/golang-rate-limiter/internal/infra/database"
	"github.com/kleytonsolinho/golang-rate-limiter/utils"
	"github.com/stretchr/testify/assert"
)

type MockRedisClient struct {
	data map[string]time.Time
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{data: make(map[string]time.Time)}
}

func (m *MockRedisClient) Create(key string, value interface{}, ttl time.Duration) error {
	m.data[key] = time.Now().Add(ttl)
	return nil
}

func (m *MockRedisClient) Exists(key string) (bool, error) {
	now := time.Now()
	if expireTime, exists := m.data[key]; exists && expireTime.After(now) {
		return true, nil
	}
	return false, nil
}

func setupRouter(rdb database.StorageStrategy) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	rateLimitWithIPPerSecond := 2
	rateLimitWithTokenPerSecond := 2
	blockDuration := 1 * time.Minute // Tempo de bloqueio após atingir o limite

	limitByIP := httprate.Limit(
		rateLimitWithIPPerSecond, // requesições/seg por IP
		1*time.Second,            // por duração
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			ip := utils.GetIP(r)
			err := rdb.Create(ip, "1", blockDuration)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		}),
	)

	limitByToken := httprate.Limit(
		rateLimitWithTokenPerSecond, // requesições/seg por Token
		1*time.Second,               // por duração
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("API_KEY")
			err := rdb.Create(token, "1", blockDuration)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		}),
		httprate.WithKeyFuncs(
			func(r *http.Request) (string, error) {
				token := r.Header.Get("API_KEY")
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
			token := r.Header.Get("API_KEY")

			// Verificar se o IP ou Token está bloqueado
			isBlockedIP, err := rdb.Exists(ip)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			isBlockedToken, err := rdb.Exists(token)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if isBlockedIP || isBlockedToken {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
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

	return r
}

func TestRateLimiter(t *testing.T) {
	mockRedis := NewMockRedisClient()
	r := setupRouter(mockRedis)

	// Testar rate limiter por Token
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("API_KEY", "test-token")
	rr := httptest.NewRecorder()

	// Enviar duas requisições para estar dentro do limite
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Enviar uma terceira requisição para atingir o limite
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)

	// Testar rate limiter por IP
	req.Header.Del("API_KEY")
	rr = httptest.NewRecorder()

	// Enviar duas requisições para estar dentro do limite
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Enviar uma terceira requisição para atingir o limite
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
}
