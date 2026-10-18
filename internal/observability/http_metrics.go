// Package observability provides low-overhead HTTP request IDs and Prometheus text metrics.
package observability

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var durationBucketSeconds = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5}

type seriesKey struct {
	method, route string
	status        int
}
type seriesValue struct {
	count      uint64
	sumSeconds float64
	buckets    []uint64
}

type HTTPMetrics struct {
	mu     sync.Mutex
	series map[seriesKey]*seriesValue
}

func NewHTTPMetrics() *HTTPMetrics { return &HTTPMetrics{series: make(map[seriesKey]*seriesValue)} }

func requestID(incoming string) string {
	incoming = strings.TrimSpace(incoming)
	if len(incoming) > 0 && len(incoming) <= 128 && !strings.ContainsAny(incoming, "\r\n\x00") {
		return incoming
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func (m *HTTPMetrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := requestID(c.GetHeader("X-Request-ID"))
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		key := seriesKey{method: c.Request.Method, route: route, status: c.Writer.Status()}
		m.mu.Lock()
		value := m.series[key]
		if value == nil {
			value = &seriesValue{buckets: make([]uint64, len(durationBucketSeconds))}
			m.series[key] = value
		}
		value.count++
		value.sumSeconds += duration
		for i, bound := range durationBucketSeconds {
			if duration <= bound {
				value.buckets[i]++
			}
		}
		m.mu.Unlock()
		log.Printf("http_request request_id=%q method=%s route=%s status=%d duration_ms=%.3f", id, c.Request.Method, route, c.Writer.Status(), duration*1000)
	}
}

func metricLabel(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return value
}

func (m *HTTPMetrics) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		m.mu.Lock()
		keys := make([]seriesKey, 0, len(m.series))
		for key := range m.series {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			a, b := keys[i], keys[j]
			if a.route != b.route {
				return a.route < b.route
			}
			if a.method != b.method {
				return a.method < b.method
			}
			return a.status < b.status
		})
		var out strings.Builder
		out.WriteString("# HELP orderflow_http_requests_total Total HTTP requests by method, registered route and status.\n")
		out.WriteString("# TYPE orderflow_http_requests_total counter\n")
		out.WriteString("# HELP orderflow_http_request_duration_seconds HTTP request duration histogram.\n")
		out.WriteString("# TYPE orderflow_http_request_duration_seconds histogram\n")
		for _, key := range keys {
			value := m.series[key]
			labels := fmt.Sprintf(`method=%q,route=%q,status=%q`, metricLabel(key.method), metricLabel(key.route), fmt.Sprint(key.status))
			fmt.Fprintf(&out, "orderflow_http_requests_total{%s} %d\n", labels, value.count)
			for i, bound := range durationBucketSeconds {
				fmt.Fprintf(&out, "orderflow_http_request_duration_seconds_bucket{%s,le=%q} %d\n", labels, fmt.Sprintf("%g", bound), value.buckets[i])
			}
			fmt.Fprintf(&out, "orderflow_http_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n", labels, value.count)
			fmt.Fprintf(&out, "orderflow_http_request_duration_seconds_sum{%s} %.9f\n", labels, value.sumSeconds)
			fmt.Fprintf(&out, "orderflow_http_request_duration_seconds_count{%s} %d\n", labels, value.count)
		}
		m.mu.Unlock()
		c.String(http.StatusOK, out.String())
	}
}
