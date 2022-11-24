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

// WithDatastoreSegment that calls all the other metrics reporters
func (m *MixedMetricsReporter) WithDatastoreSegment(table, operation string, f func() error) error {
	if m == nil {
		return f()
	}

	for _, mr := range m.MetricsReporters {
		data := mr.StartDatastoreSegment(SegmentPostgres, table, operation)
		defer mr.EndDatastoreSegment(data)
	}

	return f()
}

// WithRedisSegment with redis segment
func (m *MixedMetricsReporter) WithRedisSegment(operation string, f func() error) error {
	if m == nil {
		return f()
	}

	for _, mr := range m.MetricsReporters {
		data := mr.StartDatastoreSegment(SegmentRedis, "redis", operation)
		defer mr.EndDatastoreSegment(data)
	}

	return f()
}

// WithExternalSegment that calls all the other metrics reporters
func (m *MixedMetricsReporter) WithExternalSegment(url string, f func() error) error {
	if m == nil {
		return f()
	}

	for _, mr := range m.MetricsReporters {
		data := mr.StartExternalSegment(url)
		defer mr.EndExternalSegment(data)
	}

	return f()
}

// AddReporter to metrics reporter
func (m *MixedMetricsReporter) AddReporter(mr MetricsReporter) {
	m.MetricsReporters = append(m.MetricsReporters, mr)
}
