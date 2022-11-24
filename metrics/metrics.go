package metrics

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const DefaultAllowedMuxVars = "gameID roomType scriptName operation"

func AllowedMuxVars() []string {
	allowedString := DefaultAllowedMuxVars
	if envAllowed, ok := os.LookupEnv("MARATHON_METRICS_ALLOWEDMUXVARS"); ok {
		allowedString = envAllowed
	}
	return strings.Split(allowedString, " ")
}

func buildAPIMetricsLabels() []string {
	return append([]string{"route", "method"}, AllowedMuxVars()...)
}

var (
	// APIResponseTime summary
	APIResponseTime = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "marathon",
			Subsystem:  "api",
			Name:       "response_time_ms",
			Help:       "The response time in ms of api routes",
			Objectives: map[float64]float64{0.7: 0.02, 0.95: 0.005, 0.99: 0.001},
		},
		buildAPIMetricsLabels(),
	)

	// APIRequestsCounter counter
	APIRequestsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "marathon",
			Subsystem: "api",
			Name:      "requests_counter",
			Help:      "A counter of api requests",
		},
		buildAPIMetricsLabels(),
	)

	// APIStatusCode1XXCounter counter
	APIStatusCode1XXCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "marathon",
			Subsystem: "api",
			Name:      "requests_status_code_1xx_counter",
			Help:      "A counter of api requests with status 1xx",
		},
		buildAPIMetricsLabels(),
	)

	// APIStatusCode2XXCounter counter
	APIStatusCode2XXCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "marathon",
			Subsystem: "api",
			Name:      "requests_status_code_2xx_counter",
			Help:      "A counter of api requests with status 2xx",
		},
		buildAPIMetricsLabels(),
	)

	// APIStatusCode3XXCounter counter
	APIStatusCode3XXCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "marathon",
			Subsystem: "api",
			Name:      "requests_status_code_3xx_counter",
			Help:      "A counter of api requests with status 3xx",
		},
		buildAPIMetricsLabels(),
	)

	// APIStatusCode4XXCounter counter
	APIStatusCode4XXCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "marathon",
			Subsystem: "api",
			Name:      "requests_status_code_4xx_counter",
			Help:      "A counter of api requests with status 4xx",
		},
		buildAPIMetricsLabels(),
	)

	// APIStatusCode5XXCounter counter
	APIStatusCode5XXCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "marathon",
			Subsystem: "api",
			Name:      "requests_status_code_5xx_counter",
			Help:      "A counter of api requests with status 5xx",
		},
		buildAPIMetricsLabels(),
	)

	// WorkerRunTime summary
	WorkerRunTime = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "marathon",
			Subsystem:  "worker",
			Name:       "runtime_ms",
			Help:       "The run time of marathon scripts",
			Objectives: map[float64]float64{0.7: 0.02, 0.95: 0.005, 0.99: 0.001},
		},
		[]string{"game", "roomType", "job"},
	)

	// WorkerSuccessCounter counter
	WorkerSuccessCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "marathon",
			Subsystem: "worker",
			Name:      "success_counter",
			Help:      "The counter of successfull job runs",
		},
		[]string{"game", "roomType", "job"},
	)

	// 	WorkerErrorCounter counter
	WorkerErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "marathon",
			Subsystem: "worker",
			Name:      "error_counter",
			Help:      "The counter of failed job runs",
		},
		[]string{"game", "roomType", "job"},
	)
)

// SetupMetrics register all metrics
func init() {
	prometheus.MustRegister(
		APIResponseTime,
		APIRequestsCounter,
		WorkerRunTime,
		WorkerSuccessCounter,
		WorkerErrorCounter,
	)
	port := ":9090"
	if envPort, ok := os.LookupEnv("MARATHON_PROMETHEUS_PORT"); ok {
		port = fmt.Sprintf(":%s", envPort)
	}
	go func() {
		r := mux.NewRouter()
		r.Handle("/metrics", promhttp.Handler())

		s := &http.Server{
			Addr:           port,
			ReadTimeout:    8 * time.Second,
			WriteTimeout:   8 * time.Second,
			MaxHeaderBytes: 1 << 20,
			Handler:        r,
		}
		log.Fatal(s.ListenAndServe())
	}()
}
