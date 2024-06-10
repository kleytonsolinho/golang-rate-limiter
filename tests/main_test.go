package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kleytonsolinho/golang-rate-limiter/utils"
)

func TestRateLimit(t *testing.T) {
	rl := utils.NewRateLimiter(5 * time.Minute) // Criar um RateLimiter com um tempo de bloqueio de 5 minutos

	// Testar bloqueio por IP
	ip := "192.168.1.100"
	rl.BlockIP(ip)

	// Verificar se o IP está bloqueado
	blocked, _ := rl.IsBlocked(ip, "")
	if !blocked {
		t.Errorf("Expected IP %s to be blocked, but it's not", ip)
	}

	// Testar bloqueio por Token
	token := "my-token"
	rl.BlockToken(token)

	// Verificar se o token está bloqueado
	blocked, _ = rl.IsBlocked("", token)
	if !blocked {
		t.Errorf("Expected token %s to be blocked, but it's not", token)
	}

	// Testar se um IP bloqueado não pode fazer uma requisição
	reqIP, _ := http.NewRequest("GET", "/", nil)
	reqIP.RemoteAddr = ip
	wIP := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := utils.GetIP(r)
			blocked, _ := rl.IsBlocked(ip, "")
			if blocked {
				http.Error(w, "IP blocked", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Success!"))
	})
	r.ServeHTTP(wIP, reqIP)
	if wIP.Code != http.StatusTooManyRequests {
		t.Errorf("Expected IP to be blocked, got status code %d", wIP.Code)
	}

	// Testar se um token bloqueado não pode fazer uma requisição
	reqToken, _ := http.NewRequest("GET", "/", nil)
	reqToken.Header.Set("X-Ratelimit-Token", token)
	wToken := httptest.NewRecorder()
	rToken := chi.NewRouter()
	rToken.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Ratelimit-Token")
			blocked, _ := rl.IsBlocked("", token)
			if blocked {
				http.Error(w, "Token blocked", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	rToken.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Success!"))
	})
	rToken.ServeHTTP(wToken, reqToken)
	if wToken.Code != http.StatusTooManyRequests {
		t.Errorf("Expected token to be blocked, got status code %d", wToken.Code)
	}
}
