package proxy

import (
	"fmt"
	"io"
	"loadbalancer/internal/backend"
	"loadbalancer/internal/limiter"
	"loadbalancer/internal/store"
	"net"
	"strings"
)

func HandleConnection(clientConn net.Conn, pool *backend.ServerPool, l *limiter.RateLimiter) {
	defer clientConn.Close()

	remoteAddr := clientConn.RemoteAddr().String()
	ip := strings.Split(remoteAddr, ":")[0]

	// 1. IP Rate Limiting (Token Bucket)
	if !l.Allow(ip) {
		blockRequest(clientConn, "Rate Limit Exceeded (IP)")
		return
	}

	buffer := make([]byte, 4096)
	n, err := clientConn.Read(buffer)
	if err != nil {
		if err != io.EOF {
			fmt.Printf("Read error: %v\n", err)
		}
		return
	}

	requestData := string(buffer[:n])
	lines := strings.Split(requestData, "\n")
	firstLine := ""
	path := ""
	authHeader := ""

	if len(lines) > 0 {
		firstLine = strings.TrimSpace(lines[0])
		parts := strings.Split(firstLine, " ")
		if len(parts) >= 2 {
			path = parts[1]
		}
	}

	// Simple Header Parsing for Authorization (Case Insensitive)
	for _, line := range lines {
		// "Authorization: Bearer <token>"
		// Check case-insensitive prefix
		if len(line) > 14 && strings.EqualFold(line[:14], "Authorization:") {
			parts := strings.Split(strings.TrimSpace(line), " ")
			if len(parts) >= 3 && strings.EqualFold(parts[1], "Bearer") {
				authHeader = parts[2]
			} else if len(parts) == 2 {
				// Sometimes sent as "Authorization: <token>" (less standard but possible)
				authHeader = parts[1]
			}
		}
	}

	// 2. API Key Quota Check (Redis)
	if authHeader != "" {
		allowed, err := store.CheckQuota(authHeader)
		if err != nil {
			fmt.Printf("Redis Error: %v\n", err)
			// Fail open but log
		} else if !allowed {
			fmt.Printf("[QUOTA] API Key %s quota exceeded\n", authHeader)
			blockRequest(clientConn, "API Key Quota Exceeded")
			return
		}
	}

	fmt.Printf("[Log] İstek: %s (Path: %s)\n", firstLine, path)

	targetAddr := pool.GetPeer(path)
	fmt.Printf("[Proxy] Yönlendiriliyor -> %s\n", targetAddr)

	targetBackend := pool.GetBackend(targetAddr)
	if targetBackend != nil {
		targetBackend.IncrementRequest()
	}

	backendConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		fmt.Printf("[Hata] Backend'e ulaşılamadı (%s): %v\n", targetAddr, err)
		return
	}
	defer backendConn.Close()

	// CRITICAL FIX: Write the initial buffer to the backend!
	// Without this, the backend never sees the initial request line/headers.
	_, err = backendConn.Write(buffer[:n])
	if err != nil {
		fmt.Printf("[Hata] Backend'e yazma hatası: %v\n", err)
		return
	}

	errChan := make(chan error, 1)

	go func() {
		// Client -> Backend (remaining stream)
		_, err := io.Copy(backendConn, clientConn)
		errChan <- err
	}()

	go func() {
		// Backend -> Client
		_, err := io.Copy(clientConn, backendConn)
		errChan <- err
	}()

	<-errChan
}

func blockRequest(conn net.Conn, reason string) {
	response := "HTTP/1.1 429 Too Many Requests\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"Connection: close\r\n\r\n" +
		reason + "\n"
	conn.Write([]byte(response))
}
