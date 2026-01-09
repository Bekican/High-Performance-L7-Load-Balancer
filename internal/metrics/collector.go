package metrics

import (
	"database/sql"
	"fmt"
	"loadbalancer/internal/database"
	"sync"
	"time"
)

// Collector handles metrics collection and storage
type Collector struct {
	mu      sync.Mutex
	enabled bool
}

// MetricPoint represents a single data point
type MetricPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	BackendURL   string    `json:"backend_url,omitempty"`
	ResponseTime int64     `json:"response_time_ms,omitempty"`
	StatusCode   int       `json:"status_code,omitempty"`
	IsError      bool      `json:"is_error"`
}

// TrafficData represents traffic analytics data
type TrafficData struct {
	Timestamp string `json:"timestamp"`
	Requests  int    `json:"requests"`
	Errors    int    `json:"errors"`
}

// PerformanceData represents performance analytics
type PerformanceData struct {
	Timestamp       string  `json:"timestamp"`
	AvgResponseTime float64 `json:"avg_response_time"`
	MaxResponseTime int64   `json:"max_response_time"`
}

// Global collector instance
var DefaultCollector = &Collector{enabled: true}

// RecordRequest records a request metric
func (c *Collector) RecordRequest(backendURL string, responseTimeMs int64, statusCode int, isError bool) {
	if !c.enabled || database.DB == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := database.DB.Exec(
		"INSERT INTO metrics (backend_url, response_time_ms, status_code, is_error) VALUES (?, ?, ?, ?)",
		backendURL, responseTimeMs, statusCode, isError,
	)
	if err != nil {
		fmt.Printf("Failed to record metric: %v\n", err)
	}
}

// GetTrafficData returns traffic data for the last N hours
func GetTrafficData(hours int) ([]TrafficData, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT 
			strftime('%Y-%m-%d %H:00', timestamp) as hour,
			COUNT(*) as requests,
			SUM(CASE WHEN is_error = 1 THEN 1 ELSE 0 END) as errors
		FROM metrics
		WHERE timestamp >= datetime('now', '-' || ? || ' hours')
		GROUP BY strftime('%Y-%m-%d %H:00', timestamp)
		ORDER BY hour
	`

	rows, err := database.DB.Query(query, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []TrafficData
	for rows.Next() {
		var td TrafficData
		if err := rows.Scan(&td.Timestamp, &td.Requests, &td.Errors); err != nil {
			continue
		}
		data = append(data, td)
	}

	return data, nil
}

// GetPerformanceData returns performance metrics for the last N hours
func GetPerformanceData(hours int) ([]PerformanceData, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT 
			strftime('%Y-%m-%d %H:00', timestamp) as hour,
			AVG(response_time_ms) as avg_response,
			MAX(response_time_ms) as max_response
		FROM metrics
		WHERE timestamp >= datetime('now', '-' || ? || ' hours')
		GROUP BY strftime('%Y-%m-%d %H:00', timestamp)
		ORDER BY hour
	`

	rows, err := database.DB.Query(query, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []PerformanceData
	for rows.Next() {
		var pd PerformanceData
		var avgResp sql.NullFloat64
		var maxResp sql.NullInt64
		if err := rows.Scan(&pd.Timestamp, &avgResp, &maxResp); err != nil {
			continue
		}
		pd.AvgResponseTime = avgResp.Float64
		pd.MaxResponseTime = maxResp.Int64
		data = append(data, pd)
	}

	return data, nil
}

// GetErrorRate returns the error rate percentage
func GetErrorRate(hours int) (float64, error) {
	if database.DB == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT 
			CAST(SUM(CASE WHEN is_error = 1 THEN 1 ELSE 0 END) AS FLOAT) / 
			CAST(COUNT(*) AS FLOAT) * 100
		FROM metrics
		WHERE timestamp >= datetime('now', '-' || ? || ' hours')
	`

	var rate sql.NullFloat64
	err := database.DB.QueryRow(query, hours).Scan(&rate)
	if err != nil {
		return 0, err
	}

	return rate.Float64, nil
}

// CleanOldMetrics removes metrics older than N days (for maintenance)
func CleanOldMetrics(days int) error {
	if database.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := database.DB.Exec(
		"DELETE FROM metrics WHERE timestamp < datetime('now', '-' || ? || ' days')",
		days,
	)
	return err
}
