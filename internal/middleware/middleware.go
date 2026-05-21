package middleware

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Context keys for request-scoped values.
type ctxKey int

const (
	RequestIDKey ctxKey = iota
	UserKey
)

// RequestID generates and sets a unique request ID.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-Id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID returns the request ID from context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// Logging logs HTTP requests with structured fields.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := GetRequestID(r.Context())

		// Wrap writer to capture status code
		lw := &loggingWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(lw, r)

		duration := time.Since(start)
		slog.LogAttrs(r.Context(), slog.LevelInfo, "HTTP request",
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("query", r.URL.RawQuery),
			slog.Int("status", lw.status),
			slog.Duration("duration", duration),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)
	})
}

type loggingWriter struct {
	http.ResponseWriter
	status int
}

func (lw *loggingWriter) WriteHeader(status int) {
	lw.status = status
	lw.ResponseWriter.WriteHeader(status)
}

// CORS handles Cross-Origin Resource Sharing.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"X-Request-Id"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
}

func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	slices.Sort(cfg.AllowedOrigins)
	slices.Sort(cfg.AllowedMethods)
	slices.Sort(cfg.AllowedHeaders)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowedOrigin := ""

			if origin != "" {
				if len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
					allowedOrigin = "*"
				} else if idx, found := slices.BinarySearch(cfg.AllowedOrigins, origin); found {
					allowedOrigin = cfg.AllowedOrigins[idx]
				}
			}

			if allowedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			}

			if len(cfg.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
			}
			if len(cfg.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
			}
			if len(cfg.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))
			}
			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if cfg.MaxAge > 0 {
				ageSec := int64(cfg.MaxAge.Seconds())
				if ageSec > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.FormatInt(ageSec, 10))
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// BasicAuth provides HTTP Basic Authentication.
type BasicAuthConfig struct {
	Username string
	Password string
	Realm    string
}

func BasicAuth(cfg BasicAuthConfig) func(http.Handler) http.Handler {
	realm := cfg.Realm
	if realm == "" {
		realm = "vault-reader"
	}
	expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(cfg.Username+":"+cfg.Password))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth != expectedAuth {
				w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// TokenBucketRateLimiter implements token bucket rate limiting per IP address.
// This provides smoother rate limiting compared to fixed window algorithms.
type TokenBucketRateLimiter struct {
	mu          sync.RWMutex
	clients     map[string]*tokenBucket
	capacity    int64        // Maximum tokens in bucket
	refillRate  int64        // Tokens added per second
	refillTick  time.Duration // How often to add tokens
}

type tokenBucket struct {
	tokens      int64
	lastRefill  time.Time
	mu          sync.Mutex
}

// NewTokenBucketRateLimiter creates a token bucket rate limiter.
// capacity: max requests per burst
// refillRate: requests refilled per second
func NewTokenBucketRateLimiter(capacity, refillRate int64) *TokenBucketRateLimiter {
	rl := &TokenBucketRateLimiter{
		clients:    make(map[string]*tokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
		refillTick: 100 * time.Millisecond,
	}
	go rl.refillLoop()
	return rl
}

func (rl *TokenBucketRateLimiter) refillLoop() {
	ticker := time.NewTicker(rl.refillTick)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, tb := range rl.clients {
			tb.mu.Lock()
			elapsed := time.Since(tb.lastRefill)
			tokensToAdd := int64(float64(rl.refillRate) * elapsed.Seconds())
			tb.tokens += tokensToAdd
			if tb.tokens > rl.capacity {
				tb.tokens = rl.capacity
			}
			tb.lastRefill = time.Now()
			tb.mu.Unlock()

			// Remove idle clients
			tb.mu.Lock()
			if tb.tokens >= rl.capacity {
				delete(rl.clients, ip)
			}
			tb.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

func (rl *TokenBucketRateLimiter) Allow(ip string) bool {
	rl.mu.RLock()
	tb, exists := rl.clients[ip]
	rl.mu.RUnlock()

	if !exists {
		tb = &tokenBucket{tokens: rl.capacity, lastRefill: time.Now()}
		rl.mu.Lock()
		rl.clients[ip] = tb
		rl.mu.Unlock()
	}

	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill based on elapsed time since last request
	elapsed := time.Since(tb.lastRefill)
	tokensToAdd := int64(float64(rl.refillRate) * elapsed.Seconds())
	tb.tokens += tokensToAdd
	if tb.tokens > rl.capacity {
		tb.tokens = rl.capacity
	}
	tb.lastRefill = time.Now()

	if tb.tokens < 1 {
		return false
	}
	tb.tokens--
	return true
}

func (rl *TokenBucketRateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		if !rl.Allow(ip) {
			slog.Warn("rate limit exceeded", "ip", ip)
			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(rl.capacity, 10))
			w.Header().Set("Retry-After", "1")
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimiter is an alias for TokenBucketRateLimiter for backward compatibility.
type RateLimiter = TokenBucketRateLimiter

// NewRateLimiter creates a token bucket rate limiter with default refill rate.
// Requests per window are converted to tokens per second.
func NewRateLimiter(requests int, window time.Duration) *TokenBucketRateLimiter {
	refillRate := int64(float64(requests) / window.Seconds())
	return NewTokenBucketRateLimiter(int64(requests), refillRate)
}

// Recovery recovers from panics and logs them.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				requestID := GetRequestID(r.Context())
				slog.Error("panic recovered",
					slog.String("request_id", requestID),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Any("panic", rvr),
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Metrics tracks HTTP request metrics for Prometheus.
type Metrics struct {
	mu              sync.RWMutex
	requestsTotal   map[string]int64  // path -> count
	requestDuration map[string][]time.Duration // path -> durations
	errorsTotal     map[string]int64  // path -> error count
	activeRequests  int64
}

// NewMetrics creates a new metrics collector.
func NewMetrics() *Metrics {
	m := &Metrics{
		requestsTotal:   make(map[string]int64),
		requestDuration: make(map[string][]time.Duration),
		errorsTotal:     make(map[string]int64),
	}
	go m.cleanup()
	return m
}

func (m *Metrics) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		// Keep only last 1000 durations per path
		for path, durations := range m.requestDuration {
			if len(durations) > 1000 {
				m.requestDuration[path] = durations[len(durations)-1000:]
			}
		}
		m.mu.Unlock()
	}
}

// RecordRequest records a completed request.
func (m *Metrics) RecordRequest(method, path string, status int, duration time.Duration) {
	key := method + " " + path

	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestsTotal[key]++
	if status >= 400 {
		m.errorsTotal[key]++
	}

	durations := m.requestDuration[key]
	durations = append(durations, duration)
	if len(durations) > 1000 {
		durations = durations[len(durations)-1000:]
	}
	m.requestDuration[key] = durations
}

// IncrementActive increments active request count.
func (m *Metrics) IncrementActive() {
	m.mu.Lock()
	m.activeRequests++
	m.mu.Unlock()
}

// DecrementActive decrements active request count.
func (m *Metrics) DecrementActive() {
	m.mu.Lock()
	m.activeRequests--
	m.mu.Unlock()
}

// Handler returns the metrics handler for use in middleware chain.
func (m *Metrics) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.IncrementActive()
		defer func() {
			m.DecrementActive()
			lw, ok := w.(*loggingWriter)
			status := http.StatusOK
			if ok {
				status = lw.status
			}
			m.RecordRequest(r.Method, r.URL.Path, status, time.Since(start))
		}()
		next.ServeHTTP(w, r)
	})
}

// WritePrometheus writes metrics in Prometheus text format.
func (m *Metrics) WritePrometheus(w http.ResponseWriter) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// Request counts
	fmt.Fprintln(w, "# HELP http_requests_total Total number of HTTP requests")
	fmt.Fprintln(w, "# TYPE http_requests_total counter")
	for path, count := range m.requestsTotal {
		fmt.Fprintf(w, "http_requests_total{path=%q} %d\n", path, count)
	}

	// Error counts
	fmt.Fprintln(w, "\n# HELP http_errors_total Total number of HTTP errors")
	fmt.Fprintln(w, "# TYPE http_errors_total counter")
	for path, count := range m.errorsTotal {
		fmt.Fprintf(w, "http_errors_total{path=%q} %d\n", path, count)
	}

	// Request duration (p50, p95, p99)
	fmt.Fprintln(w, "\n# HELP http_request_duration_seconds HTTP request duration in seconds")
	fmt.Fprintln(w, "# TYPE http_request_duration_seconds histogram")
	for path, durations := range m.requestDuration {
		if len(durations) == 0 {
			continue
		}
		slices.Sort(durations)
		p50 := durations[len(durations)*50/100]
		p95 := durations[len(durations)*95/100]
		p99 := durations[len(durations)*99/100]
		fmt.Fprintf(w, `http_request_duration_seconds{path=%q,quantile="0.5"} %f`+"\n", path, p50.Seconds())
		fmt.Fprintf(w, `http_request_duration_seconds{path=%q,quantile="0.95"} %f`+"\n", path, p95.Seconds())
		fmt.Fprintf(w, `http_request_duration_seconds{path=%q,quantile="0.99"} %f`+"\n", path, p99.Seconds())
	}

	// Active requests
	fmt.Fprintf(w, "\n# HELP http_active_requests Current number of active requests\n")
	fmt.Fprintf(w, "# TYPE http_active_requests gauge\n")
	fmt.Fprintf(w, "http_active_requests %d\n", m.activeRequests)
}

// WriteRuntimeMetrics writes Go runtime metrics.
func WriteRuntimeMetrics(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	fmt.Fprintln(w, "# HELP go_goroutines Number of goroutines")
	fmt.Fprintln(w, "# TYPE go_goroutines gauge")
	fmt.Fprintf(w, "go_goroutines %d\n", runtime.NumGoroutine())

	fmt.Fprintln(w, "\n# HELP go_memstats_alloc_bytes Number of bytes allocated and in use")
	fmt.Fprintln(w, "# TYPE go_memstats_alloc_bytes gauge")
	fmt.Fprintf(w, "go_memstats_alloc_bytes %d\n", memStats.Alloc)

	fmt.Fprintln(w, "\n# HELP go_memstats_sys_bytes Number of bytes obtained from system")
	fmt.Fprintln(w, "# TYPE go_memstats_sys_bytes gauge")
	fmt.Fprintf(w, "go_memstats_sys_bytes %d\n", memStats.Sys)

	fmt.Fprintln(w, "\n# HELP go_gc_duration_seconds GC pause duration quantiles")
	fmt.Fprintln(w, "# TYPE go_gc_duration_seconds gauge")
	numGC := int(memStats.NumGC)
	if numGC > 256 {
		numGC = 256
	}
	pauses := make([]uint64, 0, numGC)
	for i := 0; i < numGC; i++ {
		pauses = append(pauses, memStats.PauseNs[i])
	}
	slices.Sort(pauses)
	if len(pauses) > 0 {
		p50 := pauses[len(pauses)*50/100]
		p95 := pauses[len(pauses)*95/100]
		p99 := pauses[len(pauses)*99/100]
		fmt.Fprintf(w, `go_gc_duration_seconds{quantile="0.5"} %f`+"\n", float64(p50)/1e9)
		fmt.Fprintf(w, `go_gc_duration_seconds{quantile="0.95"} %f`+"\n", float64(p95)/1e9)
		fmt.Fprintf(w, `go_gc_duration_seconds{quantile="0.99"} %f`+"\n", float64(p99)/1e9)
	}

	fmt.Fprintln(w, "\n# HELP go_gc_duration_seconds_count Total number of GC cycles")
	fmt.Fprintln(w, "# TYPE go_gc_duration_seconds_count counter")
	fmt.Fprintf(w, "go_gc_duration_seconds_count %d\n", memStats.NumGC)

	fmt.Fprintln(w, "\n# HELP go_info Go build information")
	fmt.Fprintln(w, "# TYPE go_info gauge")
	fmt.Fprintf(w, `go_info{version=%q} 1\n`, runtime.Version())
}

// Chain chains multiple middleware together.
func Chain(middleware ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := len(middleware) - 1; i >= 0; i-- {
			next = middleware[i](next)
		}
		return next
	}
}
