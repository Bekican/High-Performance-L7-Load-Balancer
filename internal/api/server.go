package api

import (
	"encoding/json"
	"fmt"
	"loadbalancer/internal/backend"
	"loadbalancer/internal/limiter"
	"net/http"
)

// DashboardData represents the complete dashboard statistics
type DashboardData struct {
	Backends      []backend.BackendStat `json:"backends"`
	TotalBlocked  uint64                `json:"total_blocked"`
	TotalRequests uint64                `json:"total_requests"`
	ActiveServers int                   `json:"active_servers"`
}

// BackendRequest represents a request to add or remove a backend
type BackendRequest struct {
	URL string `json:"url"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// enableCORS adds CORS headers to the response
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// corsMiddleware wraps a handler with CORS support
func corsMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}

// StartServer starts the management API server
func StartServer(pool *backend.ServerPool, l *limiter.RateLimiter, port int) {
	// Stats endpoint - GET /stats
	http.HandleFunc("/stats", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		data := DashboardData{
			Backends:      pool.GetStats(),
			TotalBlocked:  l.GetBlockedCount(),
			TotalRequests: pool.GetTotalRequests(),
			ActiveServers: pool.GetActiveCount(),
		}
		json.NewEncoder(w).Encode(data)
	}))

	// Backend management endpoint - POST/DELETE /api/backends
	http.HandleFunc("/api/backends", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case "POST":
			// Add a new backend
			var req BackendRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(APIResponse{
					Success: false,
					Message: "Invalid JSON body",
				})
				return
			}

			if req.URL == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(APIResponse{
					Success: false,
					Message: "URL is required",
				})
				return
			}

			if err := pool.AddBackend(req.URL); err != nil {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(APIResponse{
					Success: false,
					Message: err.Error(),
				})
				return
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(APIResponse{
				Success: true,
				Message: fmt.Sprintf("Backend %s added successfully", req.URL),
			})

		case "DELETE":
			// Remove a backend
			var req BackendRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(APIResponse{
					Success: false,
					Message: "Invalid JSON body",
				})
				return
			}

			if req.URL == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(APIResponse{
					Success: false,
					Message: "URL is required",
				})
				return
			}

			if err := pool.RemoveBackend(req.URL); err != nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(APIResponse{
					Success: false,
					Message: err.Error(),
				})
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(APIResponse{
				Success: true,
				Message: fmt.Sprintf("Backend %s removed successfully", req.URL),
			})

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(APIResponse{
				Success: false,
				Message: "Method not allowed",
			})
		}
	}))

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("📊 Management API çalışıyor: http://localhost:%d\n", port)
	fmt.Printf("   GET  /stats        - İstatistikler\n")
	fmt.Printf("   POST /api/backends - Sunucu ekle\n")
	fmt.Printf("   DELETE /api/backends - Sunucu kaldır\n")

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Printf("API Sunucusu hatası: %v\n", err)
	}
}
