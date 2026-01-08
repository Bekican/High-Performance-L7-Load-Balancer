package backend

import (
	"fmt"
	"loadbalancer/internal/notification"
	"loadbalancer/internal/store"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type BackendStat struct {
	URL          string `json:"url"`
	Alive        bool   `json:"alive"`
	RequestCount uint64 `json:"request_count"`
}

type ServerPool struct {
	backends   []*Backend
	counter    uint64
	routeMap   map[string]string
	mu         sync.RWMutex // Mutex for dynamic updates
	WebhookUrl string
}

// NewServerPool creates a pool. Initial URLs can come from config or empty if DB is used.
func NewServerPool(backendUrls []string, webhookUrl string) *ServerPool {
	var backends []*Backend
	for _, url := range backendUrls {
		backends = append(backends, &Backend{
			URL:   url,
			Alive: true,
		})
	}

	return &ServerPool{
		backends:   backends,
		counter:    0,
		routeMap:   make(map[string]string),
		WebhookUrl: webhookUrl,
	}
}

// AddBackend adds a new backend dynamically
func (s *ServerPool) AddBackend(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if exists
	for _, b := range s.backends {
		if b.URL == url {
			return
		}
	}
	s.backends = append(s.backends, &Backend{
		URL:   url,
		Alive: true,
	})
	fmt.Printf("[Hot Reload] Yeni backend eklendi: %s\n", url)
}

// RemoveBackend removes a backend dynamically
func (s *ServerPool) RemoveBackend(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newBackends := []*Backend{}
	for _, b := range s.backends {
		if b.URL != url {
			newBackends = append(newBackends, b)
		}
	}
	s.backends = newBackends
	fmt.Printf("[Hot Reload] Backend silindi: %s\n", url)
}

// SyncWithDB syncs the in-memory pool with the database
func (s *ServerPool) SyncWithDB() {
	dbBackends, err := store.GetBackends()
	if err != nil {
		fmt.Printf("DB Sync Hatası: %v\n", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Create a map of current backends
	currentMap := make(map[string]*Backend)
	for _, b := range s.backends {
		currentMap[b.URL] = b
	}

	// 2. Mark DB backends
	dbMap := make(map[string]bool)
	var updatedList []*Backend

	for _, url := range dbBackends {
		dbMap[url] = true
		if existing, ok := currentMap[url]; ok {
			updatedList = append(updatedList, existing)
		} else {
			// New backend found in DB
			fmt.Printf("[Sync] DB'den yeni sunucu eklendi: %s\n", url)
			updatedList = append(updatedList, &Backend{URL: url, Alive: true})
		}
	}

	s.backends = updatedList
}

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

func (s *ServerPool) AddRule(pathPrefix string, backendUrl string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routeMap[pathPrefix] = backendUrl
}

func (s *ServerPool) HealthCheckLoop() {
	for {
		s.mu.RLock()
		backends := s.backends // Copy slice header
		s.mu.RUnlock()

		for _, b := range backends {
			status := b.HealthCheck()
			// Status changed?
			if b.IsAlive() != status {
				b.SetAlive(status)
				if !status {
					msg := fmt.Sprintf("⚠️ ALERT: Backend %s is DOWN!", b.URL)
					fmt.Println(msg)
					go notification.SendSlackNotification(s.WebhookUrl, msg)
				} else {
					msg := fmt.Sprintf("✅ INFO: Backend %s is back ONLINE.", b.URL)
					fmt.Println(msg)
					go notification.SendSlackNotification(s.WebhookUrl, msg)
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}

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

func (s *ServerPool) GetPeer(path string) string {
	s.mu.RLock()
	// Read map inside lock
	target, ok := "", false
	for prefix, backendUrl := range s.routeMap {
		if strings.HasPrefix(path, prefix) {
			target = backendUrl
			ok = true
			break
		}
	}
	s.mu.RUnlock() // Release lock before searching backend list (which takes its own lock)

	if ok {
		// Find specific backend
		s.mu.RLock()
		defer s.mu.RUnlock()
		for _, b := range s.backends {
			if b.URL == target {
				if b.IsAlive() {
					return target
				} else {
					break
				}
			}
		}
	}

	peer := s.GetNextPeer()
	if peer == "" {
		return ""
	}
	return peer
}
