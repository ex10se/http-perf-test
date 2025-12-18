package rabbitmq

const (
	ExchangeName      = "go"
	QueueGolang       = "go"
	QueueSystemGolang = "system-go"
)

// GetQueueName возвращает название очереди в зависимости от типа события
func GetQueueName(isSystem bool) string {
	if isSystem {
		return QueueSystemGolang
	}
	return QueueGolang
}
