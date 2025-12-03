package rabbitmq

const (
	ExchangeName      = "golang-fasthttp-nginx"
	QueueGolang       = "golang-fasthttp-nginx"
	QueueSystemGolang = "system-golang-fasthttp-nginx"
)

// GetQueueName возвращает название очереди в зависимости от типа события
func GetQueueName(isSystem bool) string {
	if isSystem {
		return QueueSystemGolang
	}
	return QueueGolang
}
