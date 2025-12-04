# 🚀 Go Load Balancer with Rate Limiting

A production-ready **Layer 7 (Application Layer) Load Balancer** built with Go, featuring intelligent traffic distribution, rate limiting, health checks, and real-time monitoring.

[![Go Version](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 📋 Table of Contents

- [Features](#-features)
- [Architecture](#-architecture)
- [Quick Start](#-quick-start)
- [Configuration](#-configuration)
- [Docker Deployment](#-docker-deployment)
- [Testing](#-testing)
- [Monitoring](#-monitoring)
- [Project Structure](#-project-structure)
- [Troubleshooting](#-troubleshooting)

## ✨ Features

- ⚖️ **Layer 7 Load Balancing**: Round-robin distribution with intelligent path-based routing
- 🛡️ **Rate Limiting**: Token bucket algorithm (configurable: 1 req/sec, burst capacity: 3)
- 🏥 **Health Checks**: Automatic backend health monitoring with failover
- 📊 **Real-time Dashboard**: Live statistics and monitoring with auto-refresh
- 🎯 **Path-based Routing**: Custom routing rules per endpoint
- 🐳 **Docker Support**: Full containerization with Docker Compose
- 🔍 **Per-IP Tracking**: Individual rate limiting per client IP address
- 📈 **Statistics API**: RESTful API for monitoring and analytics

## 🏗️ Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌──────────────────────────────────┐
│  Load Balancer (Port 8080)       │
│  ├─ Rate Limiter                 │
│  ├─ Health Checker               │
│  └─ Path-based Router            │
└──────┬───────────────────────────┘
       │
       ├─────────────┬──────────────┐
       ▼             ▼              ▼
  ┌─────────┐  ┌─────────┐   ┌─────────┐
  │Backend 1│  │Backend 2│   │Backend N│
  │Port 80  │  │Port 80  │   │Port 80  │
  └─────────┘  └─────────┘   └─────────┘

┌──────────────────────────────────┐
│  Stats API (Port 9091)           │
│  └─ Real-time Metrics            │
└──────────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.22+** (for local development)
- **Docker & Docker Compose** (for containerized deployment)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <your-repo-url>
   cd loadbalancer
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run the load balancer**
   ```bash
   go run cmd/lb/main.go
   ```

4. **Open the dashboard**
   - Open `dashboard.html` in your browser
   - Or access stats API: `http://localhost:9091/stats`

### Expected Output

```
============================================
L7 Load Balancer Başlatıldı (:8080)
Config: config.yaml
Varsayılan Hedefler: [backend1:80 backend2:80]
Routing Kurallar: map[/admin:backend1:80 /img:backend2:80 ...]
--------------------------------------------
⚡ Rate Limiting: 1 req/sec (capacity: 3)
📊 Dashboard API: http://localhost:9091/stats
📊 Dashboard UI: dashboard.html
--------------------------------------------
⚠️  ÖNEMLİ: Rate limiting için port :8080 kullanın!
   Backend portlarına direkt istek atarsanız 
   rate limiting ÇALIŞMAZ!
============================================
İstatistik API çalışıyor: http://localhost:9091/stats
```

## ⚙️ Configuration

Edit `config.yaml` to customize your load balancer:

```yaml
port: 8080        # Load balancer port
api_port: 9091    # Stats API port

backends:
  - "backend1:80"
  - "backend2:80"

rules:
  "/admin":   "backend1:80"  # Admin routes → Backend 1
  "/img":     "backend2:80"  # Image routes → Backend 2
  "/payment": "backend1:80"  # Payment routes → Backend 1
```

### Rate Limiting Configuration

The rate limiter uses a **Token Bucket Algorithm**:

- **Rate**: 1 request/second (refill rate)
- **Capacity**: 3 tokens (burst capacity)
- **Tracking**: Per-IP address isolation

> ⚠️ **Important**: Rate limiting only works when requests go through the load balancer port (8080). Direct requests to backend ports bypass rate limiting!

## 🐳 Docker Deployment

### Using Docker Compose (Recommended)

1. **Start all services**
   ```bash
   docker-compose up -d
   ```

   This will start:
   - Load balancer (port 8080, 9091)
   - Backend 1 (internal)
   - Backend 2 (internal)

2. **View logs**
   ```bash
   docker-compose logs -f lb
   ```

3. **Stop services**
   ```bash
   docker-compose down
   ```

### Manual Docker Build

```bash
# Build the image
docker build -t go-loadbalancer .

# Run the container
docker run -p 8080:8080 -p 9091:9091 go-loadbalancer
```

### Docker Architecture

The `docker-compose.yaml` sets up:
- **go-loadbalancer**: Main load balancer service
- **backend-1**: Python HTTP server (port 80)
- **backend-2**: Python HTTP server (port 80)
- **lb-network**: Bridge network for inter-container communication

## 🧪 Testing

### Automated Test Scripts

The project includes PowerShell test scripts for comprehensive testing:

#### 1. Quick Test (Normal Traffic)
```powershell
.\test-quick.ps1
```
- Sends 10 requests with 1-second delays
- Verifies load distribution
- Checks backend health

#### 2. Aggressive Test (Rate Limiting)
```powershell
.\test-aggressive.ps1
```
- Sends 10 rapid requests (no delay)
- Tests rate limiting behavior
- Counts successful vs blocked requests
- Expected: ~3 successful, ~7 blocked (429 responses)

#### 3. Traffic Distribution Test
```powershell
.\test-traffic.ps1
```
- Tests round-robin distribution
- Verifies backend load balancing
- Displays request distribution statistics

#### 4. Rate Limit Test
```powershell
.\test-ratelimit.ps1
```
- Comprehensive rate limiting validation
- Tests burst capacity
- Verifies token refill behavior

### Manual Testing

#### Test Normal Request
```bash
curl http://localhost:8080/
```

#### Test Rate Limiting
```bash
# Send rapid requests
for i in {1..10}; do curl http://localhost:8080/; done
```

Expected: First 3 requests succeed, rest return `429 Too Many Requests`

#### Test Path-based Routing
```bash
curl http://localhost:8080/admin    # → Backend 1
curl http://localhost:8080/img      # → Backend 2
curl http://localhost:8080/payment  # → Backend 1
```

### Expected Rate Limit Response

```
HTTP/1.1 429 Too Many Requests
Content-Type: text/plain; charset=utf-8
Retry-After: 1

Rate Limit Asildi! Biraz yavasla dostum.
Limit: 1 req/sec, Capacity: 3 tokens
```

## 📊 Monitoring

### Stats API

**Endpoint**: `GET http://localhost:9091/stats`

**Response**:
```json
{
  "backends": [
    {
      "url": "backend1:80",
      "alive": true,
      "request_count": 42
    },
    {
      "url": "backend2:80",
      "alive": true,
      "request_count": 38
    }
  ],
  "total_blocked": 15
}
```

### Web Dashboard

Open `dashboard.html` in your browser for:
- 🟢 Real-time backend health status
- 📊 Request distribution charts
- 🚫 Rate limiting statistics
- ⚡ Auto-refresh every 2 seconds

### Metrics Available

- **Backend Status**: Alive/Dead for each backend
- **Request Count**: Total requests per backend
- **Blocked Requests**: Total rate-limited requests
- **Load Distribution**: Percentage distribution across backends

## 📁 Project Structure

```
loadbalancer/
├── cmd/
│   └── lb/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   └── server.go            # Stats API server
│   ├── backend/
│   │   ├── pool.go              # Backend pool management
│   │   └── server.go            # Backend server struct
│   ├── config/
│   │   └── config.go            # Configuration loader (YAML)
│   ├── limiter/
│   │   └── limiter.go           # Rate limiter (token bucket)
│   └── proxy/
│       └── handler.go           # HTTP proxy handler
├── html/
│   ├── backend1/                # Backend 1 static files
│   └── backend2/                # Backend 2 static files
├── config.yaml                  # Configuration file
├── dashboard.html               # Web monitoring dashboard
├── Dockerfile                   # Docker image definition
├── docker-compose.yaml          # Multi-container setup
├── go.mod                       # Go module dependencies
├── go.sum                       # Dependency checksums
├── test-quick.ps1              # Normal traffic test
├── test-aggressive.ps1         # Rate limiting test
├── test-traffic.ps1            # Distribution test
├── test-ratelimit.ps1          # Comprehensive rate limit test
└── README.md                    # This file
```

## 🔧 Troubleshooting

### Rate Limiting Not Working?

**Problem**: Sending requests but not seeing rate limit messages.

**Solution**: Ensure you're sending requests to the **load balancer port (8080)**, not backend ports!

```bash
# ❌ WRONG - Bypasses load balancer
curl http://localhost:80/test

# ✅ CORRECT - Goes through load balancer
curl http://localhost:8080/test
```

### Backend Server Down?

**Symptom**: Terminal shows health check warnings:
```
⚠️  UYARI: backend1:80 çöktü!
```

**Solution**: 
- Check if backend containers are running: `docker-compose ps`
- Restart backends: `docker-compose restart backend1 backend2`
- The load balancer automatically routes traffic to healthy backends

### Port Already in Use?

**Error**: `bind: address already in use`

**Solution**:
```bash
# Find process using port 8080
netstat -ano | findstr :8080

# Kill the process (Windows)
taskkill /PID <process_id> /F

# Or change port in config.yaml
```

### Docker Network Issues?

**Problem**: Backends not reachable

**Solution**:
```bash
# Recreate network
docker-compose down
docker-compose up -d

# Check network
docker network inspect loadbalancer_lb-network
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

.

## 🙏 Acknowledgments

- Built with [Go](https://go.dev/)
- Uses [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) for configuration
- Inspired by modern load balancing architectures

---

**Made with by [Bekir Can Çakmak]**
