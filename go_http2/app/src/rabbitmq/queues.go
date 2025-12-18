package rabbitmq

const (
	ExchangeName      = "go-http2"
	QueueGolang       = "go-http2"
	QueueSystemGolang = "system-go-http2"
)

// GetQueueName возвращает название очереди в зависимости от типа события
func GetQueueName(isSystem bool) string {
	if isSystem {
		return QueueSystemGolang
	}
	return QueueGolang
}
