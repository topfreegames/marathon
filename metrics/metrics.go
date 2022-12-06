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
	return append([]string{"route", "method", "status"}, AllowedMuxVars()...)
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
)

// SetupMetrics register all metrics
func init() {
	prometheus.MustRegister(
		APIResponseTime,
		APIRequestsCounter,
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
