package kafka

import (
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducer struct {
	Producer *ckafka.Producer
}

func NewKafkaProducer() KafkaProducer {
	return KafkaProducer{}
}

func (k *KafkaProducer) SetupProducer(bootstrap string) error {
	configMap := &ckafka.ConfigMap{
		"bootstrap.servers": bootstrap,
	}

	var err error
	k.Producer, err = ckafka.NewProducer(configMap)
	if err != nil {
		return err
	}
	return nil
}

func (k *KafkaProducer) Publish(msg string, topic string) error {
	message := &ckafka.Message{
		TopicPartition: ckafka.TopicPartition{Topic: &topic, Partition: ckafka.PartitionAny},
		Value:          []byte(msg),
	}

	err := k.Producer.Produce(message, nil)
	if err != nil {
		return err
	}

	return nil
}
