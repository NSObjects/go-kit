package db

import (
	"github.com/NSObjects/go-kit/config"
)

// KafkaConfig extends the base config for Kafka.
type KafkaConfig struct {
	config.KafkaConfig
	ClientID string `json:"client_id" yaml:"client_id" toml:"client_id"`
}

// Note: Kafka initialization requires github.com/IBM/sarama which adds significant
// dependencies. Projects that need Kafka should add it directly.
//
// Example usage in your project:
//
//	import "github.com/IBM/sarama"
//
//	func NewKafkaProducer(cfg KafkaConfig) (sarama.SyncProducer, error) {
//		sc := sarama.NewConfig()
//		sc.ClientID = cfg.ClientID
//		sc.Producer.RequiredAcks = sarama.WaitForAll
//		sc.Producer.Retry.Max = 3
//		sc.Producer.Return.Successes = true
//		return sarama.NewSyncProducer(cfg.Brokers, sc)
//	}
