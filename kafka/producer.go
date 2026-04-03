// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/4/2
// 描述：生产者
// *****************************************************************************

package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type Producer struct {
	config *Config
	writer *kafka.Writer
}

// NewProducer 初始化生产者
func NewProducer(config *Config) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(config.Broker),
		Topic:    config.Topic,
		Balancer: &kafka.LeastBytes{},
		Transport: &kafka.Transport{
			SASL: plain.Mechanism{
				Username: config.Username,
				Password: config.Password,
			},
			TLS: &tls.Config{
				InsecureSkipVerify: true, // 测试可用，生产换成CA
			},
		},
	}

	return &Producer{
		config: config,
		writer: writer,
	}
}

// Send 发送消息，支持 string/[]byte/struct/map/slice
// 支持 context 超时，默认 5s
func (p *Producer) Send(msg any) error {
	var val []byte
	switch m := msg.(type) {
	case string:
		val = []byte(m)
	case []byte:
		val = m
	default:
		b, err := json.Marshal(m)
		if err != nil {
			return err
		}
		val = b
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return p.writer.WriteMessages(ctx,
		kafka.Message{
			Value: val,
		},
	)
}

// Close 关闭生产者，释放连接
func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
