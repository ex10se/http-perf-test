package rabbitmq

const (
	// Exchange для маршрутизации сообщений
	ExchangeName = "golang"

	// Очередь для обычных событий
	QueueGolang = "golang"

	// Очередь для системных событий
	QueueSystemGolang = "system-golang"
)

// GetQueueName возвращает название очереди в зависимости от типа события
func GetQueueName(isSystem bool) string {
	if isSystem {
		return QueueSystemGolang
	}
	return QueueGolang
}
