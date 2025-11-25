package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/ex10se/http-peft-test/golang/models"
	"github.com/ex10se/http-peft-test/golang/rabbitmq"
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
func errorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	}); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// successResponse отправляет JSON ответ с успехом
func successResponse(w http.ResponseWriter, processed int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "SUCCESS",
		"processed": processed,
	}); err != nil {
		log.Printf("Failed to encode success response: %v", err)
	}
}

// ServeHTTP обрабатывает HTTP запрос
func (h *StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		errorResponse(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("Failed to close request body: %v", err)
		}
	}()
	// Проверяем что тело не пустое
	if len(body) == 0 {
		errorResponse(w, "Request body is required", http.StatusBadRequest)
		return
	}
	// Парсим JSON массив событий
	var events []models.StatusEvent
	if err := json.Unmarshal(body, &events); err != nil {
		log.Printf("Failed to parse JSON: %v", err)
		errorResponse(w, "Request body must be a JSON array", http.StatusBadRequest)
		return
	}
	// Проверяем что массив не пустой
	if len(events) == 0 {
		errorResponse(w, "Request body must contain at least one event", http.StatusBadRequest)
		return
	}
	// Валидируем каждое событие
	for i, event := range events {
		if err := event.Validate(); err != nil {
			log.Printf("Validation failed for event %d: %v", i, err)
			errorResponse(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	// Обрабатываем каждое событие
	var errors []map[string]string
	for _, event := range events {
		// Определяем очередь на основе is_system
		queueName := rabbitmq.GetQueueName(event.IsSystemEvent())
		// Сериализуем событие в JSON
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
			errors = append(errors, map[string]string{
				"event": event.TxID,
				"error": "Failed to serialize event",
			})
			continue
		}
		// Отправляем в RabbitMQ
		if err := h.rmqClient.Publish(queueName, eventJSON); err != nil {
			log.Printf("Failed to publish event %s: %v", event.TxID, err)
			errors = append(errors, map[string]string{
				"event": event.TxID,
				"error": err.Error(),
			})
			continue
		}
	}
	// Если были ошибки - возвращаем частичный успех
	if len(errors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMultiStatus)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "PARTIAL_SUCCESS",
			"processed": len(events) - len(errors),
			"errors":    errors,
		}); err != nil {
			log.Printf("Failed to encode partial success response: %v", err)
		}
		return
	}
	// Все события обработаны успешно
	successResponse(w, len(events))
}
