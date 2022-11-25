package model

// MixedMetricsReporter calls other metrics reporters
type MixedMetricsReporter struct {
	MetricsReporters []MetricsReporter
	Func             func(name string, f func() error) error
}

// NewMixedMetricsReporter ctor
func NewMixedMetricsReporter() *MixedMetricsReporter {
	return &MixedMetricsReporter{
		MetricsReporters: []MetricsReporter{},
	}
}

// WithSegment that calls all the other metrics reporters
func (m *MixedMetricsReporter) WithSegment(name string, f func() error) error {
	if m == nil {
		return f()
	}

	for _, mr := range m.MetricsReporters {
		data := mr.StartSegment(name)
		defer mr.EndSegment(data, name)
	}

	return f()
}
