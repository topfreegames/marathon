package messages

// KafkaMessage is the message to be sent to Kafka
type KafkaMessage struct {
	Message string
	Topic   string
}
