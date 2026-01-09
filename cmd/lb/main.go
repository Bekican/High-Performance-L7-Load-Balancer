package main

import (
	"fmt"
	"loadbalancer/internal/api"
	"loadbalancer/internal/backend"
	"loadbalancer/internal/config"
	"loadbalancer/internal/database"
	"loadbalancer/internal/limiter"
	"loadbalancer/internal/proxy"
	"net"
	"os"
	"strconv"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Hata: Config dosyası okunamadı: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	if err := database.InitDB("loadbalancer.db"); err != nil {
		fmt.Printf("Hata: Veritabanı başlatılamadı: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Create server pool
	pool := backend.NewServerPool(cfg.Backends)

	// Create rate limiter
	rateLimiter := limiter.NewLimiter(1, 3)

	// API port
	apiPort := cfg.ApiPort
	if apiPort == 0 {
		apiPort = 9091
	}

	// Start API server
	go api.StartServer(pool, rateLimiter, apiPort)

	// Start health check loop
	go pool.HealthCheckLoop()

	// Add routing rules
	for path, target := range cfg.Rules {
		pool.AddRule(path, target)
	}

	// Start TCP listener
	listenAddr := ":" + strconv.Itoa(cfg.Port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Println("Hata: Port dinlenemedi:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("============================================")
	fmt.Printf("L7 Load Balancer SaaS v2.0 (%s)\n", listenAddr)
	fmt.Printf("Config: config.yaml\n")
	fmt.Printf("Database: loadbalancer.db\n")
	fmt.Printf("Varsayılan Hedefler: %v\n", cfg.Backends)
	fmt.Printf("Routing Kurallar: %v\n", cfg.Rules)
	fmt.Println("--------------------------------------------")
	fmt.Printf("⚡ Rate Limiting: %v req/sec (capacity: %v)\n", 1.0, 3.0)
	fmt.Printf("📊 Dashboard API: http://localhost:%d\n", apiPort)
	fmt.Printf("� Auth: POST /api/auth/register, /api/auth/login\n")
	fmt.Printf("📈 Analytics: GET /api/analytics/traffic\n")
	fmt.Printf("🌐 Frontend: http://localhost:5173\n")
	fmt.Println("============================================")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go proxy.HandleConnection(conn, pool, rateLimiter)
	}
}
