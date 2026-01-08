package limiter

import (
	"testing"
	"time"
)

func TestNewLimiter(t *testing.T) {
	l := NewLimiter(1, 3)
	if l.rate != 1 {
		t.Errorf("Expected rate 1, got %f", l.rate)
	}
	if l.capacity != 3 {
		t.Errorf("Expected capacity 3, got %f", l.capacity)
	}
}

func TestAllow_Basic(t *testing.T) {
	// Rate: 1 req/sec, Capacity: 3
	l := NewLimiter(1, 3)
	ip := "127.0.0.1"

	// İlk 3 istek geçmeli
	for i := 0; i < 3; i++ {
		if !l.Allow(ip) {
			t.Errorf("Request %d should have been allowed", i+1)
		}
	}

	// 4. istek engellenmeli (token yok)
	if l.Allow(ip) {
		t.Error("Request 4 should have been blocked")
	}

	// Blocked count kontrolü
	if count := l.GetBlockedCount(); count != 1 {
		t.Errorf("Expected blocked count 1, got %d", count)
	}
}

func TestAllow_Refill(t *testing.T) {
	// Rate: 10 req/sec (hızlı dolum için), Capacity: 1
	l := NewLimiter(10, 1)
	ip := "192.168.1.1"

	// Token harca
	l.Allow(ip)

	// Hemen arkasından gelen istek engellenmeli
	if l.Allow(ip) {
		t.Error("Immediate second request should be blocked")
	}

	// 200ms bekle (10 req/sec = 100ms'de 1 token dolar)
	time.Sleep(200 * time.Millisecond)

	// Şimdi izin verilmeli
	if !l.Allow(ip) {
		t.Error("Request should be allowed after refill")
	}
}

func TestMultiIP(t *testing.T) {
	l := NewLimiter(1, 1)

	// IP A token harcar
	if !l.Allow("1.1.1.1") {
		t.Error("First IP should be allowed")
	}
	if l.Allow("1.1.1.1") {
		t.Error("First IP should be blocked")
	}

	// IP B etkilenmemeli
	if !l.Allow("2.2.2.2") {
		t.Error("Second IP should be allowed independently")
	}
}
