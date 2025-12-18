package rabbitmq

const (
	ExchangeName      = "go-echo"
	QueueGolang       = "go-echo"
	QueueSystemGolang = "system-go-echo"
)

// GetQueueName возвращает название очереди в зависимости от типа события
func GetQueueName(isSystem bool) string {
	if isSystem {
		return QueueSystemGolang
	}
	return QueueGolang
}
