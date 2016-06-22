package kafka

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/golang/glog"
)

// ReadError is cool
type consumeError struct {
	Message string
}

func (e consumeError) Error() string {
	return fmt.Sprintf("%v", e.Message)
}

// Consumer reads from the specified Kafka topic while the Messages channel is open
func Consumer(consumerConfig *ConsumerConfig, outChan chan<- string, done <-chan struct{}) {
	// Set configurations for consumer
	clusterConfig := cluster.NewConfig()
	clusterConfig.Consumer.Return.Errors = true
	clusterConfig.Group.Return.Notifications = true
	clusterConfig.Version = sarama.V0_9_0_0

	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Create consumer defined by the configurations
	consumer, consumerErr := cluster.NewConsumer(
		consumerConfig.Brokers,
		consumerConfig.ConsumerGroup,
		consumerConfig.Topics,
		clusterConfig,
	)
	if consumerErr != nil {
		glog.Errorf("Could not create consumer due to: %+v\n", consumerErr)
		return
	}
	defer consumer.Close()

	go func() {
		for err := range consumer.Errors() {
			glog.Errorf("Consumer error: %+v\n", err)
		}
	}()

	go func() {
		for notif := range consumer.Notifications() {
			glog.Info("Rebalanced: %+v\n", notif)
		}
	}()
	glog.Info("Starting kafka consumer")
	MainLoop(consumer, outChan, done)
	glog.Info("Stopped kafka consumer")
}

// MainLoop to read messages from Kafka and send them forward in the pipeline
func MainLoop(consumer *cluster.Consumer, outChan chan<- string, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return // breaks out of the for
		case msg, ok := <-consumer.Messages():
			if !ok {
				glog.Errorf("Not ok consuming from Kafka: %+v\n", msg)
				return // breaks out of the for
			}
			strMsg, err := Consume(msg)
			if err != nil {
				glog.Errorf("Error reading kafka message: %+v", err)
				continue
			}
			consumer.MarkOffset(msg, "")
			outChan <- strMsg
		}
	}
}

// Consume extracts the message from the consumer message
func Consume(kafkaMsg *sarama.ConsumerMessage) (string, error) {
	msg := string(kafkaMsg.Value)
	if msg == "" {
		return "", consumeError{"Empty message"}
	}
	glog.Info("Read message: %+v\n", msg)
	return msg, nil
}
