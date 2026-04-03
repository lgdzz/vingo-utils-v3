// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/4/2
// 描述：消费者
// *****************************************************************************

package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type Consumer struct {
	config *Config
	reader *kafka.Reader
	cancel context.CancelFunc
}

// NewConsumer 初始化消费者
func NewConsumer(config *Config, groupId string, startOffset int64) *Consumer {
	if startOffset == 0 {
		startOffset = kafka.FirstOffset
	}
	dialer := &kafka.Dialer{
		Timeout: 10 * time.Second,
		SASLMechanism: plain.Mechanism{
			Username: config.Username,
			Password: config.Password,
		},
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{config.Broker},
		Topic:       config.Topic,
		GroupID:     groupId,
		Dialer:      dialer,
		StartOffset: startOffset,
		MinBytes:    1,
		MaxBytes:    10e6, // 10MB
	})

	return &Consumer{
		config: config,
		reader: reader,
	}
}

// Start 启动消费循环，自动重连，每 10 秒重试一次
func (s *Consumer) Start(handler func(msg []byte) error) {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel // 保存 cancel，Close 时可以调用

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			err := s.consumeLoop(ctx, handler)
			log.Println("consumer stopped:", err)

			// 网络或服务断开时，等待 10 秒重试
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}
		}
	}()
}

// consumeLoop 执行一次消费循环
func (s *Consumer) consumeLoop(ctx context.Context, handler func(msg []byte) error) error {
	for {
		m, err := s.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		// 用户处理消息
		if handler != nil {
			if err := handler(m.Value); err != nil {
				log.Println("handler error:", err)
			}
		} else {
			log.Println("received message:", string(m.Value))
		}

		if err := s.reader.CommitMessages(ctx, m); err != nil {
			return err
		}
	}
}

// Close 关闭消费者，释放连接
func (s *Consumer) Close() error {
	if s.cancel != nil {
		s.cancel()
	}
	return s.reader.Close()
}
