package rabbitmq

const (
	ExchangeName      = "golang-gin-nginx"
	QueueGolang       = "golang-gin-nginx"
	QueueSystemGolang = "system-golang-gin-nginx"
)

// GetQueueName возвращает название очереди в зависимости от типа события
func GetQueueName(isSystem bool) string {
	if isSystem {
		return QueueSystemGolang
	}
	return QueueGolang
}
