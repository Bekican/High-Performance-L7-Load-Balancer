package proxy

import (
	"fmt"
	"io"
	"loadbalancer/internal/backend"
	"loadbalancer/internal/limiter"
	"net"
	"strings"
)

func HandleConnection(clientConn net.Conn, pool *backend.ServerPool, l *limiter.RateLimiter) {
	defer clientConn.Close()

	remoteAddr := clientConn.RemoteAddr().String()
	ip := strings.Split(remoteAddr, ":")[0]

	if !l.Allow(ip) {
		fmt.Printf("[RATE LIMIT] IP: %s - İstek engellendi (limit aşıldı)\n", ip)
		fmt.Printf("   Total Blocked: %d\n", l.GetBlockedCount())

		response := "HTTP/1.1 429 Too Many Requests\r\n" +
			"Content-Type: text/plain; charset=utf-8\r\n" +
			"Retry-After: 1\r\n" +
			"Connection: close\r\n\r\n" +
			"Rate Limit Asildi! Biraz yavasla dostum.\n" +
			"Limit: 1 req/sec, Capacity: 3 tokens\n"

		clientConn.Write([]byte(response))
		return
	}

	buffer := make([]byte, 4096)
	n, err := clientConn.Read(buffer)
	if err != nil {
		return
	}

	requestData := string(buffer[:n])
	lines := strings.Split(requestData, "\n")
	firstLine := ""
	path := ""

	if len(lines) > 0 {
		firstLine = strings.TrimSpace(lines[0])
		parts := strings.Split(firstLine, " ")
		if len(parts) >= 2 {
			path = parts[1]
		}
	}
	fmt.Printf("[Log] İstek: %s (Path: %s)\n", firstLine, path)

	targetAddr := pool.GetPeer(path)
	fmt.Printf("[Proxy] Yönlendiriliyor -> %s\n", targetAddr)

	// Increment request counter for traffic distribution tracking
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

	backendConn.Write(buffer[:n])

	errChan := make(chan error, 1)

	go func() {
		_, err := io.Copy(clientConn, backendConn)
		errChan <- err
	}()

	go func() {
		// Client -> Backend
		_, err := io.Copy(backendConn, clientConn)
		errChan <- err
	}()

	<-errChan
}
