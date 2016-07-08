package producer

// Config stores the required config of the Kafka producer
type Config struct {
	// List of 'host:port' strings representing the kafka brokers
	Brokers []string
}
