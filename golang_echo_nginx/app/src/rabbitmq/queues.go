package rabbitmq

const (
	ExchangeName      = "golang-echo-nginx"
	QueueGolang       = "golang-echo-nginx"
	QueueSystemGolang = "system-golang-echo-nginx"
)

// GetQueueName возвращает название очереди в зависимости от типа события
func GetQueueName(isSystem bool) string {
	if isSystem {
		return QueueSystemGolang
	}
	return QueueGolang
}
