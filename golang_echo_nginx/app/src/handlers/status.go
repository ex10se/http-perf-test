package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/ex10se/http-perf-test/golang_echo_nginx/models"
	"github.com/ex10se/http-perf-test/golang_echo_nginx/rabbitmq"
	"github.com/labstack/echo/v4"
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
func errorResponse(ctx echo.Context, message string, statusCode int) {
	ctx.Response().Header().Set("Content-Type", "application/json")
	ctx.Response().WriteHeader(statusCode)
	if err := json.NewEncoder(ctx.Response()).Encode(map[string]interface{}{
		"error": message,
	}); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// successResponse отправляет JSON ответ с успехом
func successResponse(ctx echo.Context, processed int) {
	ctx.Response().Header().Set("Content-Type", "application/json")
	ctx.Response().WriteHeader(http.StatusOK)
	if err := json.NewEncoder(ctx.Response()).Encode(map[string]interface{}{
		"status":    "SUCCESS",
		"processed": processed,
	}); err != nil {
		log.Printf("Failed to encode success response: %v", err)
	}
}

// Handle обрабатывает HTTP запрос
func (h *StatusHandler) Handle(ctx echo.Context) error {
	// Проверяем метод запроса
	if ctx.Request().Method != http.MethodPost {
		errorResponse(ctx, "Method not allowed", http.StatusMethodNotAllowed)
		return nil
	}
	// Читаем тело запроса
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil || len(body) == 0 {
		errorResponse(ctx, "Request body is required", http.StatusBadRequest)
		return nil
	}
	// Парсим JSON массив событий
	var events []models.StatusEvent
	if err := json.Unmarshal(body, &events); err != nil {
		log.Printf("Failed to parse JSON: %v", err)
		errorResponse(ctx, "Request body must be a JSON array", http.StatusBadRequest)
		return nil
	}
	// Проверяем что массив не пустой
	if len(events) == 0 {
		errorResponse(ctx, "Request body must contain at least one event", http.StatusBadRequest)
		return nil
	}
	// Валидируем каждое событие
	for i, event := range events {
		if err := event.Validate(); err != nil {
			log.Printf("Validation failed for event %d: %v", i, err)
			errorResponse(ctx, "Validation failed: "+err.Error(), http.StatusBadRequest)
			return nil
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
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(ctx.Response()).Encode(map[string]interface{}{
			"status":    "PARTIAL_SUCCESS",
			"processed": len(events) - len(errorsList),
			"errors":    errorsList,
		}); err != nil {
			log.Printf("Failed to encode partial success response: %v", err)
		}
		return nil
	}
	// Все события обработаны успешно
	successResponse(ctx, len(events))
	return nil
}
