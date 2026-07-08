// Package queue provides RabbitMQ-backed RPC command delivery for ACSGO.
//
// Commands flow: REST API → CommandRepository (PostgreSQL) → RabbitMQ queue →
// CWMP server (delivery to device on next Inform).
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/DWISSNET/acsgo/pkg/models"
	"github.com/DWISSNET/acsgo/pkg/repository"
)

const (
	// Exchange and queue names
	exchangeName  = "acsgo.commands"
	queueName     = "acsgo.rpc"
	dlxName       = "acsgo.commands.dlx"
	dlqName       = "acsgo.rpc.dead"
	retryInterval = 60 * time.Second
)

// CommandPublisher is the minimal interface used by the CWMP server.
type CommandPublisher interface {
	Publish(ctx context.Context, cmd *models.Command) error
}

// CommandMessage is the payload published to RabbitMQ.
type CommandMessage struct {
	CommandID  string            `json:"command_id"`
	DeviceID   string            `json:"device_id"`
	Type       models.CommandType `json:"type"`
	Parameters map[string]string `json:"parameters"`
	RetryCount int               `json:"retry_count"`
	Timestamp  time.Time         `json:"timestamp"`
}

// Client manages a RabbitMQ connection with auto-reconnect.
type Client struct {
	url         string
	conn        *amqp.Connection
	channel     *amqp.Channel
	commandRepo *repository.CommandRepository
	logger      *zap.Logger
	stop        chan struct{}
}

// NewClient creates a new RabbitMQ client and sets up exchanges/queues.
func NewClient(url string, commandRepo *repository.CommandRepository, logger *zap.Logger) (*Client, error) {
	c := &Client{
		url:         url,
		commandRepo: commandRepo,
		logger:      logger,
		stop:        make(chan struct{}),
	}
	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) connect() error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("rabbitmq channel: %w", err)
	}

	// Dead-letter exchange
	if err := ch.ExchangeDeclare(dlxName, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare DLX: %w", err)
	}
	if _, err := ch.QueueDeclare(dlqName, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare DLQ: %w", err)
	}
	if err := ch.QueueBind(dlqName, dlqName, dlxName, false, nil); err != nil {
		return fmt.Errorf("bind DLQ: %w", err)
	}

	// Main exchange and queue with DLX
	if err := ch.ExchangeDeclare(exchangeName, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}
	if _, err := ch.QueueDeclare(queueName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    dlxName,
		"x-dead-letter-routing-key": dlqName,
		"x-message-ttl":             int32(300_000), // 5 minutes
	}); err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}
	if err := ch.QueueBind(queueName, queueName, exchangeName, false, nil); err != nil {
		return fmt.Errorf("bind queue: %w", err)
	}

	c.conn = conn
	c.channel = ch
	c.logger.Info("RabbitMQ connected", zap.String("url", c.url))
	return nil
}

// Publish sends a command to the RabbitMQ queue.
func (c *Client) Publish(ctx context.Context, cmd *models.Command) error {
	msg := CommandMessage{
		CommandID:  cmd.ID,
		DeviceID:   cmd.DeviceID,
		Type:       cmd.CommandType,
		Parameters: cmd.Parameters,
		RetryCount: cmd.RetryCount,
		Timestamp:  time.Now(),
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal command: %w", err)
	}

	return c.channel.PublishWithContext(ctx, exchangeName, queueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Body:         body,
	})
}

// Consume starts consuming commands from the queue.
// handler is called for each message; return error to nack+retry.
func (c *Client) Consume(ctx context.Context, handler func(ctx context.Context, msg CommandMessage) error) error {
	deliveries, err := c.channel.Consume(queueName, "acsgo-consumer", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	go func() {
		for {
			select {
			case <-c.stop:
				return
			case <-ctx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					c.logger.Warn("RabbitMQ delivery channel closed, reconnecting...")
					time.Sleep(5 * time.Second)
					_ = c.connect()
					return
				}
				var msg CommandMessage
				if err := json.Unmarshal(d.Body, &msg); err != nil {
					c.logger.Error("Unmarshal command message", zap.Error(err))
					_ = d.Nack(false, false) // dead-letter
					continue
				}
				if err := handler(ctx, msg); err != nil {
					c.logger.Warn("Command handler error",
						zap.String("command_id", msg.CommandID),
						zap.Error(err),
					)
					_ = d.Nack(false, true) // requeue
				} else {
					_ = d.Ack(false)
				}
			}
		}
	}()

	return nil
}

// Close shuts down the RabbitMQ client.
func (c *Client) Close() {
	close(c.stop)
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

// NoopPublisher is a no-op publisher used when RabbitMQ is not configured.
type NoopPublisher struct{}

// Publish is a no-op.
func (n *NoopPublisher) Publish(_ context.Context, _ *models.Command) error { return nil }
