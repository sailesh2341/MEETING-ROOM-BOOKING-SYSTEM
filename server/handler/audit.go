package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sailesh-kona/meeting-room-booking-system/db/models"
)

func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 || parsed > 500 {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = parsed
	}

	rows, err := h.db.QueryContext(
		r.Context(),
		`SELECT id, user_id, action, details, created_at
		 FROM audit_logs
		 ORDER BY created_at DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list audit logs")
		return
	}
	defer rows.Close()

	logs := make([]models.AuditLog, 0)
	for rows.Next() {
		var item models.AuditLog
		var userID sql.NullInt64
		var details []byte
		err := rows.Scan(&item.ID, &userID, &item.Action, &details, &item.CreatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not read audit logs")
			return
		}
		item.Details = json.RawMessage(details)
		if userID.Valid {
			id := int(userID.Int64)
			item.UserID = &id
		}
		logs = append(logs, item)
	}

	writeJSON(w, http.StatusOK, logs)
}

func (h *Handler) logAction(ctx context.Context, userID *int, action string, details map[string]interface{}) {
	data, err := json.Marshal(details)
	if err != nil {
		data = []byte("{}")
	}
	var userValue interface{}
	if userID != nil {
		userValue = *userID
	}
	_, _ = h.db.ExecContext(
		ctx,
		`INSERT INTO audit_logs (user_id, action, details)
		 VALUES ($1, $2, $3)`,
		userValue,
		action,
		data,
	)
}
