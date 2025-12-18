package rabbitmq

const (
	ExchangeName      = "go-gin"
	QueueGolang       = "go-gin"
	QueueSystemGolang = "system-go-gin"
)

// GetQueueName возвращает название очереди в зависимости от типа события
func GetQueueName(isSystem bool) string {
	if isSystem {
		return QueueSystemGolang
	}
	return QueueGolang
}
