package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sailesh-kona/meeting-room-booking-system/config"
	"github.com/sailesh-kona/meeting-room-booking-system/server/middleware"
)

type Handler struct {
	db  *sql.DB
	cfg config.Config
}

func New(database *sql.DB, cfg config.Config) *Handler {
	return &Handler{db: database, cfg: cfg}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func readJSON(r *http.Request, value interface{}) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func getID(r *http.Request) (int, error) {
	idText := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idText)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func currentUser(r *http.Request) (middleware.User, bool) {
	return middleware.CurrentUser(r)
}

func parseTime(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("invalid time format")
}
