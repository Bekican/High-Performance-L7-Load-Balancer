package main

import (
	"fmt"
	"loadbalancer/internal/api"
	"loadbalancer/internal/backend"
	"loadbalancer/internal/config"
	"loadbalancer/internal/limiter"
	"loadbalancer/internal/proxy"
	"net"
	"os"
	"strconv"
)

func main() {

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Hata: Config dosyası okunamadı: %v\n", err)
		os.Exit(1)
	}

	pool := backend.NewServerPool(cfg.Backends)

	rateLimiter := limiter.NewLimiter(1, 3)

	apiPort := cfg.ApiPort
	if apiPort == 0 {
		apiPort = 9091
	}

	go api.StartServer(pool, rateLimiter, apiPort)

	go pool.HealthCheckLoop()

	for path, target := range cfg.Rules {
		pool.AddRule(path, target)
	}

	listenAddr := ":" + strconv.Itoa(cfg.Port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Println("Hata: Port dinlenemedi:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("============================================")
	fmt.Printf("L7 Load Balancer Başlatıldı (%s)\n", listenAddr)
	fmt.Printf("Config: config.yaml\n")
	fmt.Printf("Varsayılan Hedefler: %v\n", cfg.Backends)
	fmt.Printf("Routing Kurallar: %v\n", cfg.Rules)
	fmt.Println("--------------------------------------------")
	fmt.Printf("⚡ Rate Limiting: %v req/sec (capacity: %v)\n", 1.0, 3.0)
	fmt.Printf("📊 Dashboard API: http://localhost:%d/stats\n", apiPort)
	fmt.Printf("📊 Dashboard UI: dashboard.html\n")
	fmt.Println("--------------------------------------------")
	fmt.Printf("⚠️  ÖNEMLİ: Rate limiting için port %s kullanın!\n", listenAddr)
	fmt.Printf("   Backend portlarına direkt istek\n")
	fmt.Printf("   atarsanız rate limiting ÇALIŞMAZ!\n")
	fmt.Println("============================================")

	for {

		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go proxy.HandleConnection(conn, pool, rateLimiter)

	}
}
