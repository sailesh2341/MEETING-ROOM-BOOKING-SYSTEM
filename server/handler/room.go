package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/sailesh-kona/meeting-room-booking-system/db/models"
)

type roomRequest struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Location string `json:"location"`
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req roomRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Location = strings.TrimSpace(req.Location)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "room name is required")
		return
	}
	if req.Capacity <= 0 {
		writeError(w, http.StatusBadRequest, "capacity must be greater than zero")
		return
	}

	var room models.Room
	err := h.db.QueryRowContext(
		r.Context(),
		`INSERT INTO rooms (name, capacity, location)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, capacity, location, created_at`,
		req.Name,
		req.Capacity,
		req.Location,
	).Scan(&room.ID, &room.Name, &room.Capacity, &room.Location, &room.CreatedAt)
	if err != nil {
		writeError(w, http.StatusConflict, "room already exists")
		return
	}

	user, _ := currentUser(r)
	h.logAction(r.Context(), &user.ID, "room_created", map[string]interface{}{"room_id": room.ID})
	writeJSON(w, http.StatusCreated, room)
}

func (h *Handler) ListRooms(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	capacityText := query.Get("capacity")
	startText := query.Get("start_time")
	endText := query.Get("end_time")

	minCapacity := 0
	if capacityText != "" {
		value, err := strconv.Atoi(capacityText)
		if err != nil || value < 0 {
			writeError(w, http.StatusBadRequest, "invalid capacity")
			return
		}
		minCapacity = value
	}

	if startText != "" || endText != "" {
		h.listAvailableRooms(w, r, startText, endText, minCapacity)
		return
	}

	rows, err := h.db.QueryContext(
		r.Context(),
		`SELECT id, name, capacity, location, created_at
		 FROM rooms
		 WHERE capacity >= $1
		 ORDER BY name`,
		minCapacity,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list rooms")
		return
	}
	defer rows.Close()

	rooms, err := scanRooms(rows)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not read rooms")
		return
	}
	writeJSON(w, http.StatusOK, rooms)
}

func (h *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid room id")
		return
	}

	var room models.Room
	err = h.db.QueryRowContext(
		r.Context(),
		`SELECT id, name, capacity, location, created_at
		 FROM rooms
		 WHERE id = $1`,
		id,
	).Scan(&room.ID, &room.Name, &room.Capacity, &room.Location, &room.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "room not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not get room")
		return
	}

	writeJSON(w, http.StatusOK, room)
}

func (h *Handler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid room id")
		return
	}

	var req roomRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Location = strings.TrimSpace(req.Location)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "room name is required")
		return
	}
	if req.Capacity <= 0 {
		writeError(w, http.StatusBadRequest, "capacity must be greater than zero")
		return
	}

	var room models.Room
	err = h.db.QueryRowContext(
		r.Context(),
		`UPDATE rooms
		 SET name = $1, capacity = $2, location = $3
		 WHERE id = $4
		 RETURNING id, name, capacity, location, created_at`,
		req.Name,
		req.Capacity,
		req.Location,
		id,
	).Scan(&room.ID, &room.Name, &room.Capacity, &room.Location, &room.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "room not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusConflict, "could not update room")
		return
	}

	user, _ := currentUser(r)
	h.logAction(r.Context(), &user.ID, "room_updated", map[string]interface{}{"room_id": room.ID})
	writeJSON(w, http.StatusOK, room)
}

func (h *Handler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid room id")
		return
	}

	result, err := h.db.ExecContext(r.Context(), `DELETE FROM rooms WHERE id = $1`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete room")
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		writeError(w, http.StatusNotFound, "room not found")
		return
	}

	user, _ := currentUser(r)
	h.logAction(r.Context(), &user.ID, "room_deleted", map[string]interface{}{"room_id": id})
	writeJSON(w, http.StatusOK, map[string]string{"message": "room deleted"})
}

func (h *Handler) listAvailableRooms(w http.ResponseWriter, r *http.Request, startText string, endText string, minCapacity int) {
	startTime, err := parseTime(startText)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start time")
		return
	}
	endTime, err := parseTime(endText)
	if err != nil || !startTime.Before(endTime) {
		writeError(w, http.StatusBadRequest, "invalid end time")
		return
	}

	rows, err := h.db.QueryContext(
		r.Context(),
		`SELECT r.id, r.name, r.capacity, r.location, r.created_at
		 FROM rooms r
		 WHERE r.capacity >= $1
		   AND NOT EXISTS (
		       SELECT 1
		       FROM bookings b
		       WHERE b.room_id = r.id
		         AND b.status = 'booked'
		         AND b.start_time < $3
		         AND b.end_time > $2
		   )
		 ORDER BY r.name`,
		minCapacity,
		startTime,
		endTime,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list rooms")
		return
	}
	defer rows.Close()

	rooms, err := scanRooms(rows)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not read rooms")
		return
	}
	writeJSON(w, http.StatusOK, rooms)
}

func scanRooms(rows *sql.Rows) ([]models.Room, error) {
	rooms := make([]models.Room, 0)
	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.Name, &room.Capacity, &room.Location, &room.CreatedAt)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}
