package rabbitmq

const (
	ExchangeName      = "go-fasthttp"
	QueueGolang       = "go-fasthttp"
	QueueSystemGolang = "system-go-fasthttp"
)

// GetQueueName возвращает название очереди в зависимости от типа события
func GetQueueName(isSystem bool) string {
	if isSystem {
		return QueueSystemGolang
	}
	return QueueGolang
}
