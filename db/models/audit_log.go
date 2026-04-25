package models

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID        int             `json:"id"`
	UserID    *int            `json:"user_id"`
	Action    string          `json:"action"`
	Details   json.RawMessage `json:"details"`
	CreatedAt time.Time       `json:"created_at"`
}
