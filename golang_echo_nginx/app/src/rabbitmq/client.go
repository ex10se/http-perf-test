package rabbitmq

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client представляет подключение к RabbitMQ с автоматическим переподключением
type Client struct {
	url       string
	conn      *amqp.Connection
	channel   *amqp.Channel
	mu        sync.Mutex
	connected bool
}

// New создает новый RabbitMQ клиент
func New(url string) *Client {
	return &Client{
		url:       url,
		connected: false,
	}
}

// connect устанавливает соединение с RabbitMQ
func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var err error
	c.conn, err = amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	c.channel, err = c.conn.Channel()
	if err != nil {
		_ = c.conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}
	c.connected = true
	log.Println("Connected to RabbitMQ")
	return nil
}

// isConnected проверяет активно ли подключение
func (c *Client) isConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return false
	}
	if c.conn == nil || c.conn.IsClosed() {
		c.connected = false
		return false
	}
	if c.channel == nil {
		c.connected = false
		return false
	}
	return true
}

// ensureConnected проверяет подключение и переподключается если нужно
func (c *Client) ensureConnected() error {
	if c.isConnected() {
		return nil
	}
	// Retry логика с exponential backoff
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if err := c.connect(); err == nil {
			return nil
		}
		// Exponential backoff: 1s, 2s, 4s, 8s, 16s
		waitTime := time.Duration(1<<uint(i)) * time.Second
		log.Printf("Failed to connect to RabbitMQ, retrying in %v (attempt %d/%d)", waitTime, i+1, maxRetries)
		time.Sleep(waitTime)
	}
	return fmt.Errorf("failed to connect to RabbitMQ after %d attempts", maxRetries)
}

// compressMessage сжимает сообщение с помощью gzip
func compressMessage(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write(data); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DeclareQueues создает exchange и очереди в RabbitMQ
func (c *Client) DeclareQueues() error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	// Декларируем exchange
	err := c.channel.ExchangeDeclare(
		ExchangeName, // name
		"direct",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}
	// Декларируем очереди
	queues := []string{QueueGolang, QueueSystemGolang}
	for _, queueName := range queues {
		_, err := c.channel.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
		}
		// Привязываем очередь к exchange
		err = c.channel.QueueBind(
			queueName,    // queue name
			queueName,    // routing key
			ExchangeName, // exchange
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", queueName, err)
		}
	}
	log.Println("Queues and exchange declared successfully")
	return nil
}

// Publish отправляет сообщение в RabbitMQ с автоматическим переподключением
func (c *Client) Publish(queueName string, body []byte) error {
	// Сжимаем сообщение
	compressed, err := compressMessage(body)
	if err != nil {
		return fmt.Errorf("failed to compress message: %w", err)
	}
	// Retry логика для публикации
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if err := c.ensureConnected(); err != nil {
			return err
		}
		c.mu.Lock()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = c.channel.PublishWithContext(
			ctx,
			ExchangeName, // exchange
			queueName,    // routing key
			false,        // mandatory
			false,        // immediate
			amqp.Publishing{
				ContentType:     "application/json",
				ContentEncoding: "gzip",
				DeliveryMode:    amqp.Persistent,
				Body:            compressed,
			},
		)
		cancel()
		c.mu.Unlock()
		if err == nil {
			return nil
		}
		log.Printf("Failed to publish message: %v, retrying...", err)
		c.connected = false // Помечаем что нужно переподключение
		time.Sleep(time.Second)
	}
	return fmt.Errorf("failed to publish message after %d attempts: %w", maxRetries, err)
}

// Close закрывает соединение с RabbitMQ
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}
	c.connected = false
	log.Println("RabbitMQ connection closed")
	return nil
}
