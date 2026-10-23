package observability

import (
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

var buckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

type series struct {
	count  uint64
	sum    float64
	bucket []uint64
}

type Metrics struct {
	mu        sync.Mutex
	requests  map[string]uint64
	durations map[string]*series
	counters  map[string]uint64
	db        *sql.DB
	start     time.Time
}

func NewMetrics(db *sql.DB) *Metrics {
	return &Metrics{
		requests:  map[string]uint64{},
		durations: map[string]*series{},
		counters:  map[string]uint64{},
		db:        db,
		start:     time.Now(),
	}
}

func (m *Metrics) Inc(name string) {
	m.mu.Lock()
	m.counters[name]++
	m.mu.Unlock()
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		status := ww.Status()
		if status == 0 {
			status = http.StatusOK
		}
		route := routePattern(r)
		elapsed := time.Since(start).Seconds()
		reqKey := fmt.Sprintf(`method=%q,route=%q,status="%d"`, r.Method, route, status)
		durKey := fmt.Sprintf(`method=%q,route=%q`, r.Method, route)
		m.mu.Lock()
		m.requests[reqKey]++
		s, ok := m.durations[durKey]
		if !ok {
			s = &series{bucket: make([]uint64, len(buckets))}
			m.durations[durKey] = s
		}
		s.count++
		s.sum += elapsed
		for i, b := range buckets {
			if elapsed <= b {
				s.bucket[i]++
			}
		}
		m.mu.Unlock()
	})
}

func sortedKeys[V any](mp map[string]V) []string {
	keys := make([]string, 0, len(mp))
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b strings.Builder
		m.mu.Lock()
		b.WriteString("# TYPE http_requests_total counter\n")
		for _, k := range sortedKeys(m.requests) {
			fmt.Fprintf(&b, "http_requests_total{%s} %d\n", k, m.requests[k])
		}
		b.WriteString("# TYPE http_request_duration_seconds histogram\n")
		for _, k := range sortedKeys(m.durations) {
			s := m.durations[k]
			for i, le := range buckets {
				fmt.Fprintf(&b, "http_request_duration_seconds_bucket{%s,le=\"%s\"} %d\n", k, strconv.FormatFloat(le, 'f', -1, 64), s.bucket[i])
			}
			fmt.Fprintf(&b, "http_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n", k, s.count)
			fmt.Fprintf(&b, "http_request_duration_seconds_sum{%s} %g\n", k, s.sum)
			fmt.Fprintf(&b, "http_request_duration_seconds_count{%s} %d\n", k, s.count)
		}
		b.WriteString("# TYPE mercatus_events_total counter\n")
		for _, k := range sortedKeys(m.counters) {
			fmt.Fprintf(&b, "mercatus_events_total{event=%q} %d\n", k, m.counters[k])
		}
		m.mu.Unlock()

		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		fmt.Fprintf(&b, "# TYPE go_goroutines gauge\ngo_goroutines %d\n", runtime.NumGoroutine())
		fmt.Fprintf(&b, "# TYPE go_memstats_heap_alloc_bytes gauge\ngo_memstats_heap_alloc_bytes %d\n", ms.HeapAlloc)
		fmt.Fprintf(&b, "# TYPE process_uptime_seconds gauge\nprocess_uptime_seconds %d\n", int64(time.Since(m.start).Seconds()))
		if m.db != nil {
			st := m.db.Stats()
			fmt.Fprintf(&b, "# TYPE db_connections_open gauge\ndb_connections_open %d\n", st.OpenConnections)
			fmt.Fprintf(&b, "# TYPE db_connections_in_use gauge\ndb_connections_in_use %d\n", st.InUse)
			fmt.Fprintf(&b, "# TYPE db_connections_wait_total counter\ndb_connections_wait_total %d\n", st.WaitCount)
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(b.String()))
	})
}
