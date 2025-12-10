package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ex10se/http-perf-test/golang_gin_nginx/models"
	"github.com/ex10se/http-perf-test/golang_gin_nginx/rabbitmq"
	"github.com/gin-gonic/gin"
)

// StatusHandler обрабатывает запросы к /status/status/
type StatusHandler struct {
	rmqClient *rabbitmq.Client
}

// NewStatusHandler создает новый хэндлер
func NewStatusHandler(rmqClient *rabbitmq.Client) *StatusHandler {
	return &StatusHandler{
		rmqClient: rmqClient,
	}
}

// errorResponse отправляет JSON ответ с ошибкой
func errorResponse(ctx *gin.Context, message string, statusCode int) {
	ctx.Header("Content-Type", "application/json")
	ctx.Status(statusCode)
	if err := json.NewEncoder(ctx.Writer).Encode(map[string]interface{}{
		"error": message,
	}); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// successResponse отправляет JSON ответ с успехом
func successResponse(ctx *gin.Context, processed int) {
	ctx.Header("Content-Type", "application/json")
	ctx.Status(http.StatusOK)
	if err := json.NewEncoder(ctx.Writer).Encode(map[string]interface{}{
		"status":    "SUCCESS",
		"processed": processed,
	}); err != nil {
		log.Printf("Failed to encode success response: %v", err)
	}
}

// Handle обрабатывает HTTP запрос
func (h *StatusHandler) Handle(ctx *gin.Context) {
	// Проверяем метод запроса
	if ctx.Request.Method != http.MethodPost {
		errorResponse(ctx, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Читаем тело запроса
	body, err := ctx.GetRawData()
	if err != nil || len(body) == 0 {
		errorResponse(ctx, "Request body is required", http.StatusBadRequest)
		return
	}
	// Парсим JSON массив событий
	var events []models.StatusEvent
	if err := json.Unmarshal(body, &events); err != nil {
		log.Printf("Failed to parse JSON: %v", err)
		errorResponse(ctx, "Request body must be a JSON array", http.StatusBadRequest)
		return
	}
	// Проверяем что массив не пустой
	if len(events) == 0 {
		errorResponse(ctx, "Request body must contain at least one event", http.StatusBadRequest)
		return
	}
	// Валидируем каждое событие
	for i, event := range events {
		if err := event.Validate(); err != nil {
			log.Printf("Validation failed for event %d: %v", i, err)
			errorResponse(ctx, "Validation failed: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	// Обрабатываем каждое событие
	var errorsList []map[string]string
	for _, event := range events {
		// Определяем очередь на основе is_system
		queueName := rabbitmq.GetQueueName(event.IsSystemEvent())
		// Сериализуем событие в JSON
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
			errorsList = append(errorsList, map[string]string{
				"event": event.TxID,
				"error": "Failed to serialize event",
			})
			continue
		}
		// Отправляем в RabbitMQ
		if err := h.rmqClient.Publish(queueName, eventJSON); err != nil {
			log.Printf("Failed to publish event %s: %v", event.TxID, err)
			errorsList = append(errorsList, map[string]string{
				"event": event.TxID,
				"error": err.Error(),
			})
			continue
		}
	}
	// Если были ошибки - возвращаем частичный успех
	if len(errorsList) > 0 {
		ctx.Header("Content-Type", "application/json")
		ctx.Status(http.StatusMultiStatus)
		if err := json.NewEncoder(ctx.Writer).Encode(map[string]interface{}{
			"status":    "PARTIAL_SUCCESS",
			"processed": len(events) - len(errorsList),
			"errors":    errorsList,
		}); err != nil {
			log.Printf("Failed to encode partial success response: %v", err)
		}
		return
	}
	// Все события обработаны успешно
	successResponse(ctx, len(events))
}
