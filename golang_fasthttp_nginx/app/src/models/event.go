package models

import (
	"fmt"
)

// ErrorData представляет информацию об ошибке в событии
type ErrorData struct {
	Code    *string `json:"code,omitempty"`
	Message *string `json:"message,omitempty"`
}

// TrackData содержит метаданные трекинга события
type TrackData struct {
	Priority *int `json:"priority,omitempty"`
	IsSystem bool `json:"is_system"`
}

// StatusEvent представляет событие статуса для обработки
type StatusEvent struct {
	State     string     `json:"state"`
	Error     *ErrorData `json:"error,omitempty"`
	TrackData *TrackData `json:"trackData,omitempty"`
	UpdatedAt string     `json:"updatedAt"`
	TxID      string     `json:"txId"`
	Email     *string    `json:"email,omitempty"`
	ChannelID *string    `json:"channel_id,omitempty"`
	Channel   *string    `json:"channel,omitempty"`
}

// Validate проверяет обязательные поля события
func (e *StatusEvent) Validate() error {
	if e.State == "" {
		return fmt.Errorf("field 'state' is required")
	}
	if e.UpdatedAt == "" {
		return fmt.Errorf("field 'updatedAt' is required")
	}
	if e.TxID == "" {
		return fmt.Errorf("field 'txId' is required")
	}
	return nil
}

// IsSystemEvent возвращает true если событие системное
func (e *StatusEvent) IsSystemEvent() bool {
	if e.TrackData != nil {
		return e.TrackData.IsSystem
	}
	return false
}
