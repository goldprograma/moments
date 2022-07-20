package pkg

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusMonitor struct {
	ServiceName string //监控服务的名称

	Uptime   *prometheus.CounterVec
	ReqCount *prometheus.CounterVec

	ReqDuration   *prometheus.HistogramVec
	ReqSizeBytes  *prometheus.SummaryVec
	RespSizeBytes *prometheus.SummaryVec

	ExcludeRegexStatus   string
	ExcludeRegexEndpoint string
	ExcludeRegexMethod   string
}

func NewPrometheusMonitor(namespace, serviceName string) *PrometheusMonitor {
	var promes = &PrometheusMonitor{ServiceName: serviceName}
	labels := []string{"status", "endpoint", "method", "micro_name"}

	promes.Uptime = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "uptime",
			Help:      "HTTP service uptime.",
		}, nil,
	)

	promes.ReqCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_request_count_total",
			Help:      "Total number of HTTP requests made.",
		}, labels,
	)

	promes.ReqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latencies in seconds.",
		}, labels,
	)

	promes.ReqSizeBytes = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Name:      "http_request_size_bytes",
			Help:      "HTTP request sizes in bytes.",
		}, labels,
	)

	promes.RespSizeBytes = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Name:      "http_response_size_bytes",
			Help:      "HTTP request sizes in bytes.",
		}, labels,
	)
	prometheus.MustRegister(promes.Uptime, promes.ReqCount, promes.ReqDuration, promes.ReqSizeBytes, promes.RespSizeBytes)
	go promes.recordUptime()
	return promes
}

// recordUptime increases service uptime per second.
func (p *PrometheusMonitor) recordUptime() {
	for range time.Tick(time.Second) {
		p.Uptime.WithLabelValues().Inc()
	}
}

// calcRequestSize returns the size of request object.
func calcRequestSize(r *http.Request) float64 {
	size := 0
	if r.URL != nil {
		size = len(r.URL.String())
	}

	size += len(r.Method)
	size += len(r.Proto)

	for name, values := range r.Header {
		size += len(name)
		for _, value := range values {
			size += len(value)
		}
	}
	size += len(r.Host)

	// r.Form and r.MultipartForm are assumed to be included in r.URL.
	if r.ContentLength != -1 {
		size += int(r.ContentLength)
	}
	return float64(size)
}

// PromOpts represents the Prometheus middleware Options.
// It was used for filtering labels with regex.
// type PromOpts struct {

// }

// checkLabel returns the match result of labels.
// Return true if regex-pattern compiles failed.
func (p *PrometheusMonitor) checkLabel(label, pattern string) bool {
	if pattern == "" {
		return true
	}

	matched, err := regexp.MatchString(label, pattern)
	if err != nil {
		return true
	}
	return !matched
}

// PromMiddleware returns a gin.HandlerFunc for exporting some Web metrics
func (p *PrometheusMonitor) PromMiddleware() gin.HandlerFunc {
	// make sure promOpts is not nil

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := fmt.Sprintf("%d", c.Writer.Status())
		endpoint := c.Request.URL.Path
		method := c.Request.Method

		lvs := []string{status, endpoint, method, p.ServiceName}

		isOk := p.checkLabel(status, p.ExcludeRegexStatus) &&
			p.checkLabel(endpoint, p.ExcludeRegexEndpoint) &&
			p.checkLabel(method, p.ExcludeRegexMethod)

		if !isOk {
			return
		}

		p.ReqCount.WithLabelValues(lvs...).Inc()
		p.ReqDuration.WithLabelValues(lvs...).Observe(time.Since(start).Seconds())
		p.ReqSizeBytes.WithLabelValues(lvs...).Observe(calcRequestSize(c.Request))
		p.RespSizeBytes.WithLabelValues(lvs...).Observe(float64(c.Writer.Size()))
	}
}

// PromHandler wrappers the standard http.Handler to gin.HandlerFunc
func (p *PrometheusMonitor) PromHandler(handler http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
