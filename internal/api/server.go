package api

import (
	"context"
	"encoding/json"
	"fmt"
	"loadbalancer/internal/auth"
	"loadbalancer/internal/backend"
	"loadbalancer/internal/database"
	"loadbalancer/internal/limiter"
	"loadbalancer/internal/metrics"
	"loadbalancer/internal/plans"
	"net/http"
	"strings"
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

// AuthRequest represents login/register request
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents login response with token
type AuthResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Token   string        `json:"token,omitempty"`
	User    *UserResponse `json:"user,omitempty"`
}

// UserResponse represents user data in responses
type UserResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Plan  string `json:"plan"`
}

// ContextKey for user ID in context
type ContextKey string

const UserContextKey ContextKey = "user"

// enableCORS adds CORS headers to the response
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// corsMiddleware wraps a handler with CORS support
func corsMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}

// authMiddleware validates JWT token and adds user to context
func authMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"success":false,"message":"Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"success":false,"message":"Invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, `{"success":false,"message":"Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Add user claims to context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		handler(w, r.WithContext(ctx))
	})
}

// getUserFromContext extracts user claims from request context
func getUserFromContext(r *http.Request) *auth.Claims {
	if claims, ok := r.Context().Value(UserContextKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}

// StartServer starts the management API server
func StartServer(pool *backend.ServerPool, l *limiter.RateLimiter, port int) {
	// =====================
	// Public Endpoints
	// =====================

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

	// =====================
	// Auth Endpoints
	// =====================

	// Register - POST /api/auth/register
	http.HandleFunc("/api/auth/register", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Method not allowed"})
			return
		}

		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Invalid JSON"})
			return
		}

		if req.Email == "" || req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Email and password required"})
			return
		}

		if len(req.Password) < 6 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Password must be at least 6 characters"})
			return
		}

		// Check if user exists
		existing, _ := database.GetUserByEmail(req.Email)
		if existing != nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Email already registered"})
			return
		}

		// Hash password
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Failed to process password"})
			return
		}

		// Create user
		user, err := database.CreateUser(req.Email, hash)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Failed to create user"})
			return
		}

		// Generate token
		token, err := auth.GenerateToken(user.ID, user.Email, user.Plan)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Failed to generate token"})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: true,
			Token:   token,
			User: &UserResponse{
				ID:    user.ID,
				Email: user.Email,
				Plan:  user.Plan,
			},
		})
	}))

	// Login - POST /api/auth/login
	http.HandleFunc("/api/auth/login", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Method not allowed"})
			return
		}

		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Invalid JSON"})
			return
		}

		// Get user
		user, err := database.GetUserByEmail(req.Email)
		if err != nil || user == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Invalid email or password"})
			return
		}

		// Check password
		if !auth.CheckPassword(user.PasswordHash, req.Password) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Invalid email or password"})
			return
		}

		// Generate token
		token, err := auth.GenerateToken(user.ID, user.Email, user.Plan)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Failed to generate token"})
			return
		}

		json.NewEncoder(w).Encode(AuthResponse{
			Success: true,
			Token:   token,
			User: &UserResponse{
				ID:    user.ID,
				Email: user.Email,
				Plan:  user.Plan,
			},
		})
	}))

	// Get current user - GET /api/auth/me (protected)
	http.HandleFunc("/api/auth/me", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := getUserFromContext(r)
		if claims == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(AuthResponse{Success: false, Message: "Unauthorized"})
			return
		}

		json.NewEncoder(w).Encode(AuthResponse{
			Success: true,
			User: &UserResponse{
				ID:    claims.UserID,
				Email: claims.Email,
				Plan:  claims.Plan,
			},
		})
	}))

	// =====================
	// Backend Management
	// =====================

	// Backend management endpoint - POST/DELETE /api/backends
	http.HandleFunc("/api/backends", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case "POST":
			var req BackendRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Invalid JSON body"})
				return
			}

			if req.URL == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "URL is required"})
				return
			}

			if err := pool.AddBackend(req.URL); err != nil {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(APIResponse{Success: false, Message: err.Error()})
				return
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(APIResponse{
				Success: true,
				Message: fmt.Sprintf("Backend %s added successfully", req.URL),
			})

		case "DELETE":
			var req BackendRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Invalid JSON body"})
				return
			}

			if req.URL == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "URL is required"})
				return
			}

			if err := pool.RemoveBackend(req.URL); err != nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(APIResponse{Success: false, Message: err.Error()})
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(APIResponse{
				Success: true,
				Message: fmt.Sprintf("Backend %s removed successfully", req.URL),
			})

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Method not allowed"})
		}
	}))

	// =====================
	// Analytics Endpoints
	// =====================

	// Traffic analytics - GET /api/analytics/traffic
	http.HandleFunc("/api/analytics/traffic", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := getUserFromContext(r)
		if !plans.HasAnalytics(claims.Plan) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Analytics requires Pro or Enterprise plan"})
			return
		}

		data, err := metrics.GetTrafficData(24)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Failed to get traffic data"})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    data,
		})
	}))

	// Performance analytics - GET /api/analytics/performance
	http.HandleFunc("/api/analytics/performance", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := getUserFromContext(r)
		if !plans.HasAnalytics(claims.Plan) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Analytics requires Pro or Enterprise plan"})
			return
		}

		data, err := metrics.GetPerformanceData(24)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(APIResponse{Success: false, Message: "Failed to get performance data"})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    data,
		})
	}))

	// =====================
	// Plans Endpoint
	// =====================

	// Get all plans - GET /api/plans
	http.HandleFunc("/api/plans", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"plans":   plans.GetAllPlans(),
		})
	}))

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("📊 Management API çalışıyor: http://localhost:%d\n", port)
	fmt.Printf("   Auth:      POST /api/auth/register, /api/auth/login\n")
	fmt.Printf("   Backends:  POST/DELETE /api/backends\n")
	fmt.Printf("   Analytics: GET /api/analytics/traffic, /performance\n")
	fmt.Printf("   Plans:     GET /api/plans\n")

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Printf("API Sunucusu hatası: %v\n", err)
	}
}
