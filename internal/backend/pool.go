package backend

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// BackendStat represents statistics for a single backend server
type BackendStat struct {
	URL          string `json:"url"`
	Alive        bool   `json:"alive"`
	RequestCount uint64 `json:"request_count"`
}

// ServerPool manages a collection of backend servers with thread-safe operations
type ServerPool struct {
	backends []*Backend
	counter  uint64
	routeMap map[string]string
	mu       sync.RWMutex // Protects backends slice and routeMap
}

// NewServerPool creates a new server pool with the given backend URLs
func NewServerPool(backendUrls []string) *ServerPool {
	var backends []*Backend
	for _, url := range backendUrls {
		backends = append(backends, &Backend{
			URL:   url,
			Alive: true,
		})
	}

	return &ServerPool{
		backends: backends,
		counter:  0,
		routeMap: make(map[string]string),
	}
}

// AddBackend dynamically adds a new backend server to the pool
func (s *ServerPool) AddBackend(url string) error {
	if url == "" {
		return errors.New("backend URL cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate
	for _, b := range s.backends {
		if b.URL == url {
			return errors.New("backend already exists")
		}
	}

	newBackend := &Backend{
		URL:   url,
		Alive: true,
	}
	s.backends = append(s.backends, newBackend)
	fmt.Printf("✅ Backend eklendi: %s\n", url)
	return nil
}

// RemoveBackend dynamically removes a backend server from the pool
func (s *ServerPool) RemoveBackend(url string) error {
	if url == "" {
		return errors.New("backend URL cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, b := range s.backends {
		if b.URL == url {
			// Remove by swapping with last element and truncating
			s.backends[i] = s.backends[len(s.backends)-1]
			s.backends = s.backends[:len(s.backends)-1]
			fmt.Printf("🗑️  Backend kaldırıldı: %s\n", url)
			return nil
		}
	}

	return errors.New("backend not found")
}

// GetBackend returns a backend by its URL (thread-safe)
func (s *ServerPool) GetBackend(url string) *Backend {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, b := range s.backends {
		if b.URL == url {
			return b
		}
	}
	return nil
}

// GetStats returns statistics for all backends (thread-safe)
func (s *ServerPool) GetStats() []BackendStat {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var stats []BackendStat
	for _, b := range s.backends {
		stats = append(stats, BackendStat{
			URL:          b.URL,
			Alive:        b.IsAlive(),
			RequestCount: b.GetRequestCount(),
		})
	}
	return stats
}

// GetTotalRequests returns the sum of all backend request counts
func (s *ServerPool) GetTotalRequests() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var total uint64
	for _, b := range s.backends {
		total += b.GetRequestCount()
	}
	return total
}

// GetActiveCount returns the number of alive backends
func (s *ServerPool) GetActiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, b := range s.backends {
		if b.IsAlive() {
			count++
		}
	}
	return count
}

// AddRule adds a routing rule for path-based routing
func (s *ServerPool) AddRule(pathPrefix string, backendUrl string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routeMap[pathPrefix] = backendUrl
}

// HealthCheckLoop continuously checks the health of all backends
func (s *ServerPool) HealthCheckLoop() {
	for {
		s.mu.RLock()
		backends := make([]*Backend, len(s.backends))
		copy(backends, s.backends)
		s.mu.RUnlock()

		for _, b := range backends {
			status := b.HealthCheck()
			b.SetAlive(status)
			if !status {
				fmt.Printf("⚠️  UYARI: %s çöktü!\n", b.URL)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// GetNextPeer returns the next available backend using round-robin (thread-safe)
func (s *ServerPool) GetNextPeer() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	next := atomic.AddUint64(&s.counter, 1)
	l := len(s.backends)
	if l == 0 {
		return ""
	}

	for i := 0; i < l; i++ {
		idx := (int(next) + i) % l
		if s.backends[idx].IsAlive() {
			return s.backends[idx].URL
		}
	}
	return ""
}

// GetPeer returns a backend for the given path using routing rules or round-robin
func (s *ServerPool) GetPeer(path string) string {
	s.mu.RLock()

	for prefix, backendUrl := range s.routeMap {
		if strings.HasPrefix(path, prefix) {
			for _, b := range s.backends {
				if b.URL == backendUrl {
					if b.IsAlive() {
						s.mu.RUnlock()
						return backendUrl
					} else {
						break
					}
				}
			}
		}
	}
	s.mu.RUnlock()

	// Fall back to round-robin
	return s.GetNextPeer()
}
