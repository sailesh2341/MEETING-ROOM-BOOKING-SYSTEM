package handler

import (
	"encoding/json"
	"meeting-room-booking/db"
	"meeting-room-booking/models"
	"meeting-room-booking/utils"
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	json.NewDecoder(r.Body).Decode(&user)

	// Check user in DB (dummy query)
	row := db.DB.QueryRow("SELECT id, password, role FROM users WHERE username = $1", user.Username)
	err := row.Scan(&user.ID, &user.Password, &user.Role)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, _ := utils.GenerateJWT(user.ID, user.Role)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
