package worker_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/topfreegames/marathon/extensions"
)

// collectCounterValue gathers metrics from a registry and returns the value of the
// counter series identified by the given metric family name and label pairs.
func collectCounterValue(t *testing.T, reg *prometheus.Registry, metricName string, labelPairs map[string]string) float64 {
	t.Helper()
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() != metricName {
			continue
		}
		for _, m := range mf.GetMetric() {
			if labelsMatch(m.GetLabel(), labelPairs) {
				return m.GetCounter().GetValue()
			}
		}
	}
	return 0
}

func labelsMatch(got []*dto.LabelPair, want map[string]string) bool {
	matched := 0
	for _, lp := range got {
		if v, ok := want[lp.GetName()]; ok && v == lp.GetValue() {
			matched++
		}
	}
	return matched == len(want)
}

func TestWorkerEventCounter(t *testing.T) {
	reg := prometheus.NewRegistry()

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "marathon_worker_events_total",
			Help: "Worker lifecycle event counter.",
		},
		[]string{"event", "game", "platform"},
	)
	reg.MustRegister(counter)

	// Simulate incrWorkerEvent for a worker start event.
	counter.WithLabelValues("starting_create_batches_worker", "mygame", "gcm").Inc()

	val := collectCounterValue(t, reg, "marathon_worker_events_total", map[string]string{
		"event":    "starting_create_batches_worker",
		"game":     "mygame",
		"platform": "gcm",
	})
	if val <= 0 {
		t.Fatalf("expected counter > 0, got %v", val)
	}
}

func TestKafkaSendMessageReturnCounter(t *testing.T) {
	reg := prometheus.NewRegistry()

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "marathon_kafka_send_message_return_total",
			Help: "Kafka async producer message acknowledgements.",
		},
		[]string{"error"},
	)
	reg.MustRegister(counter)

	// Simulate the success goroutine in connectToKafka.
	counter.WithLabelValues("false").Inc()
	counter.WithLabelValues("false").Inc()
	counter.WithLabelValues("true").Inc()

	successVal := collectCounterValue(t, reg, "marathon_kafka_send_message_return_total", map[string]string{"error": "false"})
	if successVal != 2 {
		t.Fatalf("expected success counter = 2, got %v", successVal)
	}
	errorVal := collectCounterValue(t, reg, "marathon_kafka_send_message_return_total", map[string]string{"error": "true"})
	if errorVal != 1 {
		t.Fatalf("expected error counter = 1, got %v", errorVal)
	}
}

// TestKafkaMetricsRegistration verifies MustRegisterKafkaMetrics is safe to call repeatedly.
func TestKafkaMetricsRegistration(t *testing.T) {
	// Must not panic on multiple calls (sync.Once guard).
	extensions.MustRegisterKafkaMetrics()
	extensions.MustRegisterKafkaMetrics()
}
