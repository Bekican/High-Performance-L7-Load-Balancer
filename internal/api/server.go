package api

import (
	"encoding/json"
	"fmt"
	"loadbalancer/internal/backend"
	"loadbalancer/internal/limiter"
	"net/http"
)

type DashboardData struct {
	Backends     []backend.BackendStat `json:"backends"`
	TotalBlocked uint64                `json:"total_blocked"`
}

func StartServer(pool *backend.ServerPool, l *limiter.RateLimiter, port int) {
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		data := DashboardData{
			Backends:     pool.GetStats(),
			TotalBlocked: l.GetBlockedCount(),
		}
		json.NewEncoder(w).Encode(data)
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("İstatistik API çalışıyor: http://localhost:%d/stats\n", port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Printf("API Sunucusu hatası: %v\n", err)
	}
}
