package rabbitmq

const (
	// Exchange для маршрутизации сообщений
	ExchangeName = "golang_nginx_http2"

	// Очередь для обычных событий
	QueueGolang = "golang_nginx_http2"

	// Очередь для системных событий
	QueueSystemGolang = "system-golang-nginx_http2"
)

// GetQueueName возвращает название очереди в зависимости от типа события
func GetQueueName(isSystem bool) string {
	if isSystem {
		return QueueSystemGolang
	}
	return QueueGolang
}
