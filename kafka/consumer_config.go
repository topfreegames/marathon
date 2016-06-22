package kafka

// ConsumerConfig stores configuration required to initialize the Kafka
// consumers
type ConsumerConfig struct {
	// REQUIRED: The consumer group to aggregate all the consumers that belongs
	// to the same cell.
	ConsumerGroup string
	// REQUIRED: Array of strings containing the kafka topics that each consumer
	// from the same cell should subscribe to.
	Topics []string
	// REQUIRED: List of 'host:port' strings representing the kafka brokers
	// that belongs to the cell.
	Brokers []string
}
