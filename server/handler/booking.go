package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/lib/pq"
	"github.com/sailesh-kona/meeting-room-booking-system/db/models"
)

type bookingRequest struct {
	RoomID    int    `json:"room_id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var req bookingRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	startTime, endTime, ok := h.validateBookingRequest(w, req)
	if !ok {
		return
	}

	user, _ := currentUser(r)
	tx, err := h.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not start booking")
		return
	}
	defer tx.Rollback()

	var roomExists bool
	err = tx.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM rooms WHERE id = $1)`, req.RoomID).Scan(&roomExists)
	if err != nil || !roomExists {
		writeError(w, http.StatusNotFound, "room not found")
		return
	}

	var hasConflict bool
	err = tx.QueryRowContext(
		r.Context(),
		`SELECT EXISTS (
		    SELECT 1
		    FROM bookings
		    WHERE room_id = $1
		      AND status = 'booked'
		      AND start_time < $3
		      AND end_time > $2
		)`,
		req.RoomID,
		startTime,
		endTime,
	).Scan(&hasConflict)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not check room availability")
		return
	}
	if hasConflict {
		writeError(w, http.StatusConflict, "room is already booked for this time")
		return
	}

	var booking models.Booking
	err = tx.QueryRowContext(
		r.Context(),
		`INSERT INTO bookings (room_id, user_id, start_time, end_time, status)
		 VALUES ($1, $2, $3, $4, 'booked')
		 RETURNING id, room_id, user_id, start_time, end_time, status, created_at, updated_at`,
		req.RoomID,
		user.ID,
		startTime,
		endTime,
	).Scan(&booking.ID, &booking.RoomID, &booking.UserID, &booking.StartTime, &booking.EndTime, &booking.Status, &booking.CreatedAt, &booking.UpdatedAt)
	if isBookingConflict(err) {
		writeError(w, http.StatusConflict, "room is already booked for this time")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create booking")
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save booking")
		return
	}

	h.logAction(r.Context(), &user.ID, "booking_created", map[string]interface{}{"booking_id": booking.ID, "room_id": booking.RoomID})
	writeJSON(w, http.StatusCreated, booking)
}

func (h *Handler) UpdateBooking(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid booking id")
		return
	}

	var req bookingRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	startTime, endTime, ok := h.validateBookingRequest(w, req)
	if !ok {
		return
	}

	user, _ := currentUser(r)
	tx, err := h.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update booking")
		return
	}
	defer tx.Rollback()

	query := `UPDATE bookings
	          SET room_id = $1, start_time = $2, end_time = $3, updated_at = now()
	          WHERE id = $4 AND status = 'booked'`
	args := []interface{}{req.RoomID, startTime, endTime, id}
	if user.Role != "admin" {
		query += ` AND user_id = $5`
		args = append(args, user.ID)
	}
	query += ` RETURNING id, room_id, user_id, start_time, end_time, status, created_at, updated_at`

	var booking models.Booking
	err = tx.QueryRowContext(r.Context(), query, args...).Scan(&booking.ID, &booking.RoomID, &booking.UserID, &booking.StartTime, &booking.EndTime, &booking.Status, &booking.CreatedAt, &booking.UpdatedAt)
	if isBookingConflict(err) {
		writeError(w, http.StatusConflict, "room is already booked for this time")
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "booking not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update booking")
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save booking")
		return
	}

	h.logAction(r.Context(), &user.ID, "booking_updated", map[string]interface{}{"booking_id": booking.ID, "room_id": booking.RoomID})
	writeJSON(w, http.StatusOK, booking)
}

func (h *Handler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid booking id")
		return
	}

	user, _ := currentUser(r)
	query := `UPDATE bookings
	          SET status = 'cancelled', updated_at = now()
	          WHERE id = $1 AND status = 'booked'`
	args := []interface{}{id}
	if user.Role != "admin" {
		query += ` AND user_id = $2`
		args = append(args, user.ID)
	}
	query += ` RETURNING id, room_id, user_id, start_time, end_time, status, created_at, updated_at`

	var booking models.Booking
	err = h.db.QueryRowContext(r.Context(), query, args...).Scan(&booking.ID, &booking.RoomID, &booking.UserID, &booking.StartTime, &booking.EndTime, &booking.Status, &booking.CreatedAt, &booking.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "booking not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not cancel booking")
		return
	}

	h.logAction(r.Context(), &user.ID, "booking_cancelled", map[string]interface{}{"booking_id": booking.ID, "room_id": booking.RoomID})
	writeJSON(w, http.StatusOK, booking)
}

func (h *Handler) ListMyBookings(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r)
	rows, err := h.db.QueryContext(
		r.Context(),
		`SELECT id, room_id, user_id, start_time, end_time, status, created_at, updated_at
		 FROM bookings
		 WHERE user_id = $1
		 ORDER BY start_time DESC`,
		user.ID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list bookings")
		return
	}
	defer rows.Close()

	bookings, err := scanBookings(rows)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not read bookings")
		return
	}
	writeJSON(w, http.StatusOK, bookings)
}

func (h *Handler) ListBookings(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status != "" && status != "booked" && status != "cancelled" {
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}

	query := `SELECT id, room_id, user_id, start_time, end_time, status, created_at, updated_at
	          FROM bookings`
	args := []interface{}{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	query += ` ORDER BY start_time DESC`

	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list bookings")
		return
	}
	defer rows.Close()

	bookings, err := scanBookings(rows)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not read bookings")
		return
	}
	writeJSON(w, http.StatusOK, bookings)
}

func (h *Handler) validateBookingRequest(w http.ResponseWriter, req bookingRequest) (time.Time, time.Time, bool) {
	if req.RoomID <= 0 {
		writeError(w, http.StatusBadRequest, "room id is required")
		return time.Time{}, time.Time{}, false
	}
	startTime, err := parseTime(req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start time")
		return time.Time{}, time.Time{}, false
	}
	endTime, err := parseTime(req.EndTime)
	if err != nil || !startTime.Before(endTime) {
		writeError(w, http.StatusBadRequest, "invalid end time")
		return time.Time{}, time.Time{}, false
	}
	return startTime, endTime, true
}

func scanBookings(rows *sql.Rows) ([]models.Booking, error) {
	bookings := make([]models.Booking, 0)
	for rows.Next() {
		var booking models.Booking
		err := rows.Scan(&booking.ID, &booking.RoomID, &booking.UserID, &booking.StartTime, &booking.EndTime, &booking.Status, &booking.CreatedAt, &booking.UpdatedAt)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, rows.Err()
}

func isBookingConflict(err error) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23P01" || pqErr.Constraint == "no_booking_overlap"
	}
	return false
}
