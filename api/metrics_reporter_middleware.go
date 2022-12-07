package api

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/topfreegames/marathon/metrics"
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

// Serve executes the middleware logic.
// Implementation based on: https://github.com/labstack/echo/v4-contrib/blob/master/prometheus/prometheus.go#L394
func (m *MetricsReporterMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		err := next(c)
		status := c.Response().Status
		if err != nil {
			var httpError *echo.HTTPError
			if errors.As(err, &httpError) {
				status = httpError.Code
			}
			if status == 0 || status == http.StatusOK {
				status = http.StatusInternalServerError
			}
		}

		elapsed := float64(time.Since(start)) / float64(time.Second)

		labels := m.prometheusLabelsFromRequest(c.Request())
		labels["status"] = strconv.Itoa(status)

		metrics.APIResponseTime.With(labels).Observe(elapsed)
		metrics.APIRequestsCounter.With(labels).Inc()

		return err
	}
}

func (m *MetricsReporterMiddleware) prometheusLabelsFromRequest(r *http.Request) prometheus.Labels {
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
