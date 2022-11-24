package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	tracing "github.com/topfreegames/go-extensions-tracing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/topfreegames/marathon/metrics"
	"github.com/topfreegames/marathon/model"
)

// MetricsReporterMiddleware handles logging
type MetricsReporterMiddleware struct {
	App            *Application
	Next           http.Handler
	AllowedMuxVars map[string]bool
}

// NewMetricsReporterMiddleware creates a new metrics reporter middleware
func NewMetricsReporterMiddleware(a *Application) *MetricsReporterMiddleware {
	allowedMuxVarsSl := metrics.AllowedMuxVars()
	allowedMuxVarsMap := map[string]bool{}
	for _, v := range allowedMuxVarsSl {
		allowedMuxVarsMap[v] = true
	}
	m := &MetricsReporterMiddleware{App: a, AllowedMuxVars: allowedMuxVarsMap}
	return m
}

type contextKey string

// MetricsReporterKey is the key used to store the metrics reporter in the context
const metricsReporterKey = contextKey("metricsReporter")

// MixedMetricsReporter is responsible for storing the mixed metrics reporter in the context
func newContextWithMetricsReporter(ctx context.Context, mr *model.MixedMetricsReporter) context.Context {
	c := context.WithValue(ctx, metricsReporterKey, mr)
	return c
}

func trace(ctx context.Context, operationName string, tags opentracing.Tags, f func()) {
	var parent opentracing.SpanContext
	if span := opentracing.SpanFromContext(ctx); span != nil {
		parent = span.Context()
	}
	reference := opentracing.ChildOf(parent)
	if tags == nil {
		tags = opentracing.Tags{}
	}
	tags["span.kind"] = "server"
	span := opentracing.StartSpan(operationName, reference, tags)
	defer span.Finish()
	defer tracing.LogPanic(span)
	f()
}

func (m *MetricsReporterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	trace(r.Context(), "MetricsReporterMiddleware", nil, func() {
		ctx := newContextWithMetricsReporter(r.Context(), model.NewMixedMetricsReporter())

		// Call the next middleware/handler in chain
		startTime := time.Now()
		m.Next.ServeHTTP(w, r.WithContext(ctx))
		defer func() {
			labels := m.PrometheusLabelsFromRequest(r)
			timeUsed := float64(time.Since(startTime).Nanoseconds() / (1000 * 1000))
			statusCode := r.Response.StatusCode
			if statusCode >= 100 && statusCode < 200 {
				metrics.APIStatusCode1XXCounter.With(labels).Inc()
			} else if statusCode >= 200 && statusCode < 300 {
				metrics.APIStatusCode2XXCounter.With(labels).Inc()
			} else if statusCode >= 300 && statusCode < 400 {
				metrics.APIStatusCode3XXCounter.With(labels).Inc()
			} else if statusCode >= 400 && statusCode < 500 {
				metrics.APIStatusCode4XXCounter.With(labels).Inc()
			} else if statusCode >= 500 {
				metrics.APIStatusCode5XXCounter.With(labels).Inc()
			}
			metrics.APIResponseTime.With(labels).Observe(timeUsed)
			metrics.APIRequestsCounter.With(labels).Inc()
		}()
	})
}

func (m MetricsReporterMiddleware) PrometheusLabelsFromRequest(r *http.Request) prometheus.Labels {
	//keep only allowedMuxVars
	labels := prometheus.Labels{}
	for k := range m.AllowedMuxVars {
		labels[k] = ""
	}
	for k, v := range mux.Vars(r) {
		if _, ok := m.AllowedMuxVars[k]; ok {
			labels[k] = v
		}
	}
	labels["method"] = r.Method
	path := r.URL.Path
	//try to get template path with {vars} instead of the real requested paths
	//example: /games/sniper3d becomes /games/{game}
	if route := mux.CurrentRoute(r); route != nil {
		if template, err := route.GetPathTemplate(); err == nil {
			path = template
		}
	}
	labels["route"] = path
	return labels
}
