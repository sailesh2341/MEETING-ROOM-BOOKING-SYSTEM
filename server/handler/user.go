package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/sailesh-kona/meeting-room-booking-system/db/models"
	"github.com/sailesh-kona/meeting-room-booking-system/utils"
)

type registerRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Role      string `json:"role"`
	AdminCode string `json:"admin_code"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type roleRequest struct {
	Role string `json:"role"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Role = strings.TrimSpace(req.Role)
	if req.Role == "" {
		req.Role = "user"
	}
	if req.Username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}
	if err := utils.ValidatePassword(req.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Role != "user" && req.Role != "admin" {
		writeError(w, http.StatusBadRequest, "invalid role")
		return
	}
	if req.Role == "admin" && (h.cfg.AdminCode == "" || req.AdminCode != h.cfg.AdminCode) {
		writeError(w, http.StatusForbidden, "admin code is invalid")
		return
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	var user models.User
	err = h.db.QueryRowContext(
		r.Context(),
		`INSERT INTO users (username, password, role)
		 VALUES ($1, $2, $3)
		 RETURNING id, username, role, created_at`,
		req.Username,
		passwordHash,
		req.Role,
	).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)
	if err != nil {
		writeError(w, http.StatusConflict, "username already exists")
		return
	}

	h.logAction(r.Context(), &user.ID, "user_registered", map[string]interface{}{"role": user.Role})
	writeJSON(w, http.StatusCreated, user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var user models.User
	err := h.db.QueryRowContext(
		r.Context(),
		`SELECT id, username, password, role, created_at
		 FROM users
		 WHERE username = $1`,
		strings.TrimSpace(req.Username),
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not login")
		return
	}
	if !utils.CheckPassword(req.Password, user.Password) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := utils.GenerateToken(h.cfg.JWTSecret, user, h.cfg.TokenHours)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create token")
		return
	}

	h.logAction(r.Context(), &user.ID, "user_login", map[string]interface{}{"username": user.Username})
	writeJSON(w, http.StatusOK, map[string]interface{}{"token": token, "user": user})
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(
		r.Context(),
		`SELECT id, username, role, created_at
		 FROM users
		 ORDER BY id`,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list users")
		return
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not read users")
			return
		}
		users = append(users, user)
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req roleRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Role != "user" && req.Role != "admin" {
		writeError(w, http.StatusBadRequest, "invalid role")
		return
	}

	var user models.User
	err = h.db.QueryRowContext(
		r.Context(),
		`UPDATE users
		 SET role = $1
		 WHERE id = $2
		 RETURNING id, username, role, created_at`,
		req.Role,
		id,
	).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update user")
		return
	}

	admin, _ := currentUser(r)
	h.logAction(r.Context(), &admin.ID, "user_role_updated", map[string]interface{}{"user_id": user.ID, "role": user.Role})
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	result, err := h.db.ExecContext(r.Context(), `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete user")
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	admin, _ := currentUser(r)
	h.logAction(r.Context(), &admin.ID, "user_deleted", map[string]interface{}{"user_id": id})
	writeJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}
