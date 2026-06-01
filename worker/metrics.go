package worker

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	CreateBatchesWorkerStart     = "starting_create_batches_worker"
	CreateBatchesWorkerCompleted = "completed_create_batches_worker"
	CreateBatchesWorkerError     = "error_create_batches_worker"

	CsvSplitWorkerStart     = "starting_csv_split_worker"
	CsvSplitWorkerCompleted = "completed_csv_split_worker"
	CsvSplitWorkerError     = "error_csv_split_worker"

	DirectWorkerStart     = "starting_direct_part"
	DirectWorkerCompleted = "completed_direct_worker"
	DirectWorkerError     = "error_direct_worker"

	JobCompletedWorkerStart     = "starting_job_completed_worker"
	JobCompletedWorkerCompleted = "completed_job_completed_worker"
	JobCompletedWorkerError     = "error_job_completed_worker"

	ProcessBatchWorkerStart     = "starting_process_batch_worker"
	ProcessBatchWorkerCompleted = "completed_process_batch_worker"
	ProcessBatchWorkerError     = "error_process_batch_worker"

	ResumeJobWorkerStart     = "starting_resume_job_worker"
	ResumeJobWorkerCompleted = "completed_resume_job_worker"
	ResumeJobWorkerError     = "error_resume_job_worker"

	GetCsvFromS3Timing   = "get_csv_from_s3"
	GetUsersFromDbTiming = "get_from_pg"
)

var (
	registerMetricsOnce sync.Once

	// workerCounter counts worker lifecycle events (start/completed/error and csv_job_part).
	// Label names match the statsd tag keys previously passed via job.Labels():
	//   game=<app name>, platform=<service>
	// The metric name uses underscores; the DD OpenMetrics check maps marathon_* → marathon.*
	workerCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "marathon_worker_events_total",
			Help: "Worker lifecycle event counter (start/completed/error) per worker type and job labels.",
		},
		[]string{"event", "game", "platform"},
	)

	// workerDuration observes timing operations (get_csv_from_s3, get_from_pg, get_csv_batch_from_pg, save_control_group).
	workerDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "marathon_worker_duration_milliseconds",
			Help:    "Worker operation duration in milliseconds.",
			Buckets: []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000, 30000},
		},
		[]string{"operation", "game", "platform"},
	)
)

// MustRegisterMetrics registers Marathon worker Prometheus collectors. Safe to call more than once.
func MustRegisterMetrics() {
	registerMetricsOnce.Do(func() {
		prometheus.MustRegister(workerCounter)
		prometheus.MustRegister(workerDuration)
	})
}

// incrWorkerEvent increments the worker event counter for the given event name and job labels.
// labels is the slice returned by job.Labels() — each element has the form "key:value".
func incrWorkerEvent(event string, labels []string) {
	game, platform := parseJobLabels(labels)
	workerCounter.WithLabelValues(event, game, platform).Inc()
}

// observeWorkerDuration records an operation duration for the given timing name and job labels.
func observeWorkerDuration(operation string, d time.Duration, labels []string) {
	game, platform := parseJobLabels(labels)
	workerDuration.WithLabelValues(operation, game, platform).Observe(float64(d.Milliseconds()))
}

// parseJobLabels extracts game and platform from a Labels() slice of "key:value" strings.
func parseJobLabels(labels []string) (game, platform string) {
	for _, l := range labels {
		if len(l) > 5 && l[:5] == "game:" {
			game = l[5:]
		} else if len(l) > 9 && l[:9] == "platform:" {
			platform = l[9:]
		}
	}
	return
}
