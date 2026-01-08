package main

import (
	"crypto/tls"
	"fmt"
	"loadbalancer/internal/api"
	"loadbalancer/internal/backend"
	"loadbalancer/internal/config"
	"loadbalancer/internal/limiter"
	"loadbalancer/internal/proxy"
	"loadbalancer/internal/store"
	"net"
	"os"
	"strconv"

	"golang.org/x/crypto/acme/autocert"
)

func main() {
	// Initialize Stores
	if err := store.InitSQLite("loadbalancer.db"); err != nil {
		fmt.Printf("Warning: SQLite Init Failed: %v\n", err)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	if err := store.InitRedis(redisAddr); err != nil {
		fmt.Printf("Warning: Redis Init Failed: %v\n", err)
	}

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Hata: Config dosyası okunamadı: %v\n", err)
		os.Exit(1)
	}

	pool := backend.NewServerPool(cfg.Backends, cfg.WebhookUrl)

	dbBackends, _ := store.GetBackends()
	if len(dbBackends) == 0 {
		for _, u := range cfg.Backends {
			store.AddBackend(u)
		}
	}
	pool.SyncWithDB()

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

	var listener net.Listener

	// SSL Termination with Autocert
	if cfg.SSLDomain != "" {
		fmt.Printf("🔒 SSL/TLS Enabled for domain: %s\n", cfg.SSLDomain)
		m := &autocert.Manager{
			Cache:      autocert.DirCache("certs"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.SSLDomain),
		}

		// Listen on 443 for HTTPS
		// Note: Requires root privileges usually
		tlsConfig := m.TLSConfig()
		var err error
		listener, err = tls.Listen("tcp", ":443", tlsConfig)
		if err != nil {
			fmt.Printf("Failed to bind :443 for SSL: %v\n", err)
			os.Exit(1)
		}

		// Also start HTTP redirect server on port 80 or config port
		go func() {
			httpListener, _ := net.Listen("tcp", listenAddr)
			// Simple redirect or just plain http parallel
			fmt.Printf("HTTP listening on %s (Parallel)\n", listenAddr)
			for {
				conn, err := httpListener.Accept()
				if err == nil {
					go proxy.HandleConnection(conn, pool, rateLimiter)
				}
			}
		}()

	} else {
		// Standard HTTP
		var err error
		listener, err = net.Listen("tcp", listenAddr)
		if err != nil {
			fmt.Println("Hata: Port dinlenemedi:", err)
			os.Exit(1)
		}
	}

	defer listener.Close()

	fmt.Println("============================================")
	if cfg.SSLDomain != "" {
		fmt.Printf("L7 Load Balancer Başlatıldı (HTTPS: :443)\n")
	} else {
		fmt.Printf("L7 Load Balancer Başlatıldı (%s)\n", listenAddr)
	}
	fmt.Printf("Config: config.yaml + SQLite\n")
	fmt.Printf("Redis: %s\n", redisAddr)
	fmt.Println("--------------------------------------------")
	fmt.Printf("📊 Dashboard API: http://localhost:%d/stats\n", apiPort)
	fmt.Printf("📊 Dashboard UI: dashboard.html\n")
	fmt.Println("============================================")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go proxy.HandleConnection(conn, pool, rateLimiter)
	}
}
