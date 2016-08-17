package consumer

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// FIXME: Try to use a better way to define log level (do the same in the entire project)

// consumeError is an error generated during data consumption
type consumeError struct {
	Message string
}

func (e consumeError) Error() string {
	return fmt.Sprintf("%v", e.Message)
}

// Consumer reads from the specified Kafka topic while the Messages channel is open
func Consumer(l zap.Logger, config *viper.Viper, configRoot, app, service string, outChan chan<- string, doneChan <-chan struct{}) error {
	// Set configurations for consumer
	clusterConfig := cluster.NewConfig()
	clusterConfig.Consumer.Return.Errors = true
	clusterConfig.Group.Return.Notifications = true
	clusterConfig.Version = sarama.V0_9_0_0
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	brokers := config.GetStringSlice(fmt.Sprintf("%s.consumer.brokers", configRoot))
	consumerGroup := config.GetString(fmt.Sprintf("%s.consumer.consumergroup", configRoot))
	topicTemplate := config.GetString(fmt.Sprintf("%s.consumer.topicTemplate", configRoot))
	topics := []string{fmt.Sprintf(topicTemplate, app, service)}
	l.Warn(
		"Create consumer group",
		zap.String("brokers", fmt.Sprintf("%+v", brokers)),
		zap.String("consumerGroup", consumerGroup),
		zap.String("topicTemplate", topicTemplate),
		zap.Object("topics", topics),
		zap.Object("clusterConfig", clusterConfig),
	)

	// Create consumer defined by the configurations
	consumer, err := cluster.NewConsumer(brokers, consumerGroup, topics, clusterConfig)
	if err != nil {
		l.Error("Could not create consumer", zap.String("error", err.Error()))
		return err
	}
	defer consumer.Close()

	go func() {
		for err := range consumer.Errors() {
			l.Error("Consumer error", zap.String("error", err.Error()))
		}
	}()

	go func() {
		for notif := range consumer.Notifications() {
			l.Info("Rebalanced", zap.String("", fmt.Sprintf("%+v", notif)))
		}
	}()
	l.Info("Starting kafka consumer")
	MainLoop(l, consumer, outChan, doneChan)
	l.Info("Stopped kafka consumer")
	return nil
}

// MainLoop to read messages from Kafka and send them forward in the pipeline
func MainLoop(l zap.Logger, consumer *cluster.Consumer, outChan chan<- string, doneChan <-chan struct{}) {
	for {
		select {
		case <-doneChan:
			return // breaks out of the for
		case msg, ok := <-consumer.Messages():
			if !ok {
				l.Error("Not ok consuming from Kafka", zap.String("msg", fmt.Sprintf("%+v", msg)))
				return // breaks out of the for
			}
			strMsg, err := Consume(l, msg)
			if err != nil {
				l.Error("Error reading kafka message", zap.Error(err))
				continue
			}
			// FIXME: Is it the rigth place to mark offset?
			consumer.MarkOffset(msg, "")
			outChan <- strMsg
		}
	}
}

// Consume extracts the message from the consumer message
func Consume(l zap.Logger, kafkaMsg *sarama.ConsumerMessage) (string, error) {
	l.Info("Consume message", zap.String("msg", fmt.Sprintf("%+v", kafkaMsg)))
	msg := string(kafkaMsg.Value)
	if msg == "" {
		return "", consumeError{"Empty message"}
	}
	l.Info("Consumed message", zap.String("msg", msg))
	return msg, nil
}
