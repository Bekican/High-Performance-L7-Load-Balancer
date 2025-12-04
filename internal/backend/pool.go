package backend

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type BackendStat struct {
	URL          string `json:"url"`
	Alive        bool   `json:"alive"`
	RequestCount uint64 `json:"request_count"`
}

type ServerPool struct {
	backends []*Backend
	counter  uint64
	routeMap map[string]string
}

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

func (s *ServerPool) GetBackend(url string) *Backend {
	for _, b := range s.backends {
		if b.URL == url {
			return b
		}
	}
	return nil
}

func (s *ServerPool) GetStats() []BackendStat {
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
	s.routeMap[pathPrefix] = backendUrl
}

func (s *ServerPool) HealthCheckLoop() {
	for {
		for _, b := range s.backends {
			status := b.HealthCheck()
			b.SetAlive(status)
			if !status {
				fmt.Printf("⚠️  UYARI: %s çöktü!\n", b.URL)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func (s *ServerPool) GetNextPeer() string {
	next := atomic.AddUint64(&s.counter, 1)
	l := len(s.backends)
	for i := 0; i < l; i++ {
		idx := (int(next) + i) % l
		if s.backends[idx].IsAlive() {
			return s.backends[idx].URL
		}
	}
	return ""
}

func (s *ServerPool) GetPeer(path string) string {
	for prefix, backendUrl := range s.routeMap {
		if strings.HasPrefix(path, prefix) {
			for _, b := range s.backends {
				if b.URL == backendUrl {
					if b.IsAlive() {
						return backendUrl
					} else {
						break
					}
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
