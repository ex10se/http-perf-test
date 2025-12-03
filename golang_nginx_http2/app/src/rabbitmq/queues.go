package rabbitmq

const (
	ExchangeName      = "golang-nginx-http2"
	QueueGolang       = "golang-nginx-http2"
	QueueSystemGolang = "system-golang-nginx-http2"
)

// GetQueueName возвращает название очереди в зависимости от типа события
func GetQueueName(isSystem bool) string {
	if isSystem {
		return QueueSystemGolang
	}
	return QueueGolang
}
