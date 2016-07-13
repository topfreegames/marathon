package consumer

import (
	"fmt"
	"os"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

func getLogLevel() zap.Level {
	var level = zap.WarnLevel
	var environment = os.Getenv("ENV")
	if environment == "test" {
		level = zap.FatalLevel
	}
	return level
}

// Logger is the consumer logger
var Logger = zap.NewJSON(getLogLevel())

// consumeError is an error generated during data consumption
type consumeError struct {
	Message string
}

func (e consumeError) Error() string {
	return fmt.Sprintf("%v", e.Message)
}

// Consumer reads from the specified Kafka topic while the Messages channel is open
func Consumer(config *viper.Viper, outChan chan<- string, done <-chan struct{}) {
	// Set configurations for consumer
	clusterConfig := cluster.NewConfig()
	clusterConfig.Consumer.Return.Errors = true
	clusterConfig.Group.Return.Notifications = true
	clusterConfig.Version = sarama.V0_9_0_0

	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Create consumer defined by the configurations
	consumer, consumerErr := cluster.NewConsumer(
		config.GetStringSlice("workers.consumer.brokers"),
		config.GetString("workers.consumer.consumergroup"),
		config.GetStringSlice("workers.consumer.topics"),
		clusterConfig,
	)
	if consumerErr != nil {
		Logger.Error(
			"Could not create consumer",
			zap.String("error", consumerErr.Error()),
		)
		return
	}
	defer consumer.Close()

	go func() {
		for err := range consumer.Errors() {
			Logger.Error(
				"Consumer error",
				zap.String("error", err.Error()),
			)
		}
	}()

	go func() {
		for notif := range consumer.Notifications() {
			Logger.Info(
				"Rebalanced",
				zap.String("", fmt.Sprintf("%+v", notif)),
			)
		}
	}()
	Logger.Info("Starting kafka consumer")
	MainLoop(consumer, outChan, done)
	Logger.Info("Stopped kafka consumer")
}

// MainLoop to read messages from Kafka and send them forward in the pipeline
func MainLoop(consumer *cluster.Consumer, outChan chan<- string, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return // breaks out of the for
		case msg, ok := <-consumer.Messages():
			if !ok {
				Logger.Error("Not ok consuming from Kafka", zap.String("msg", fmt.Sprintf("%+v", msg)))
				return // breaks out of the for
			}
			strMsg, err := Consume(msg)
			if err != nil {
				Logger.Error("Error reading kafka message", zap.Error(err))
				continue
			}
			consumer.MarkOffset(msg, "")
			outChan <- strMsg
		}
	}
}

// Consume extracts the message from the consumer message
func Consume(kafkaMsg *sarama.ConsumerMessage) (string, error) {
	Logger.Info(
		"Consume message",
		zap.String("msg", fmt.Sprintf("%+v", kafkaMsg)),
	)
	msg := string(kafkaMsg.Value)
	if msg == "" {
		return "", consumeError{"Empty message"}
	}
	Logger.Info(
		"Consumed message",
		zap.String("msg", msg),
	)
	return msg, nil
}
