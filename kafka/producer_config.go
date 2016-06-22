package kafka

// ProducerConfig stores the required config of the Kafka producer
type ProducerConfig struct {
	// List of 'host:port' strings representing the kafka brokers
	Brokers []string
}
