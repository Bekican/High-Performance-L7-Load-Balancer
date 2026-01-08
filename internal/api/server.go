package api

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"loadbalancer/internal/backend"
	"loadbalancer/internal/limiter"
	"loadbalancer/internal/store"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type DashboardData struct {
	Backends     []backend.BackendStat `json:"backends"`
	TotalBlocked uint64                `json:"total_blocked"`
}

type AddBackendRequest struct {
	URL string `json:"url"`
}

type CreateKeyRequest struct {
	Quota int `json:"quota"`
}

type CreateKeyResponse struct {
	Key   string `json:"key"`
	Quota int    `json:"quota"`
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow CORS OPTIONS without auth
		if r.Method == "OPTIONS" {
			next(w, r)
			return
		}

		user, pass, ok := r.BasicAuth()
		// Default credentials for demo: admin / secret
		// In production, these should come from Env or DB
		envUser := os.Getenv("ADMIN_USER")
		if envUser == "" {
			envUser = "admin"
		}
		envPass := os.Getenv("ADMIN_PASS")
		if envPass == "" {
			envPass = "secret"
		}

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(envUser)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(envPass)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func StartServer(pool *backend.ServerPool, l *limiter.RateLimiter, port int) {
	// Enable CORS for all handlers
	enableCors := func(w http.ResponseWriter) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	}

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		w.Header().Set("Content-Type", "application/json")

		data := DashboardData{
			Backends:     pool.GetStats(),
			TotalBlocked: l.GetBlockedCount(),
		}
		json.NewEncoder(w).Encode(data)
	})

	// Add Backend API (Protected)
	http.HandleFunc("/api/admin/backends", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method == "OPTIONS" {
			return
		}

		if r.Method == "POST" {
			var req AddBackendRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// 1. Save to DB
			if err := store.AddBackend(req.URL); err != nil {
				http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// 2. Hot Reload
			pool.AddBackend(req.URL)

			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, "Backend added: %s", req.URL)
			return
		}

		if r.Method == "DELETE" {
			// Expecting ?url=...
			url := r.URL.Query().Get("url")
			if url == "" {
				http.Error(w, "Missing url parameter", http.StatusBadRequest)
				return
			}

			// 1. Remove from DB
			if err := store.RemoveBackend(url); err != nil {
				http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// 2. Hot Reload
			pool.RemoveBackend(url)

			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Backend removed: %s", url)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// API Key Management API (Protected)
	http.HandleFunc("/api/admin/keys", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method == "OPTIONS" {
			return
		}

		if r.Method == "POST" {
			var req CreateKeyRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if req.Quota <= 0 {
				req.Quota = 1000 // Default
			}

			apiKey := uuid.New().String()

			// 1. Save to Redis (Quota)
			if err := store.SetQuota(apiKey, req.Quota); err != nil {
				http.Error(w, "Redis Error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CreateKeyResponse{
				Key:   apiKey,
				Quota: req.Quota,
			})
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("İstatistik ve Admin API çalışıyor: http://localhost:%d/stats\n", port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Printf("API Sunucusu hatası: %v\n", err)
	}
}
