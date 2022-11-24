package model

type MetricsReporter interface {
	StartSegment(string) map[string]interface{}
	EndSegment(map[string]interface{}, string)

	StartDatastoreSegment(datastore, collection, operation string) map[string]interface{}
	EndDatastoreSegment(map[string]interface{})

	StartExternalSegment(string) map[string]interface{}
	EndExternalSegment(map[string]interface{})
}
