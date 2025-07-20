package models

import "time"

type Message struct {
	ID          string                 `json:"id" validate:"required,uuid4"`
	Timestamp   time.Time              `json:"timestamp" validate:"required"`
	Source      string                 `json:"source" validate:"required"`
	Destination string                 `json:"destination" validate:"required"`
	Protocol    string                 `json:"protocol" validate:"required,oneof=modbus dnp3 mqtt http"`
	Payload     map[string]interface{} `json:"payload" validate:"required"`
	Metadata    map[string]string      `json:"metadata"`
}

type MessageBatch struct {
	Messages []Message `json:"messages" validate:"required,dive"`
}
