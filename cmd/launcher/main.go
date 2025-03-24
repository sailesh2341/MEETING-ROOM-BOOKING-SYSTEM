package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"meeting-room-booking/config"
	"meeting-room-booking/db"
	"meeting-room-booking/routes"

	"github.com/gorilla/mux"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Connect to the database
	db.InitDB()
	defer db.CloseDB()

	// Initialize router
	r := mux.NewRouter()

	// Register routes
	routes.RegisterUserRoutes(r)
	routes.RegisterRoomRoutes(r)
	routes.RegisterBookingRoutes(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("🚀 Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// package main

// import (
// 	"database/sql"
// 	"encoding/json"
// 	"log"
// 	"net/http"
// 	"time"

// 	"github.com/dgrijalva/jwt-go"
// 	"github.com/gorilla/mux"
// 	_ "github.com/lib/pq"
// )

// var db *sql.DB
// var jwtKey = []byte("secret_key")

// type User struct {
// 	ID       int    `json:"id"`
// 	Username string `json:"username"`
// 	Password string `json:"password"`
// 	Role     string `json:"role"`
// }

// type Room struct {
// 	ID   int    `json:"id"`
// 	Name string `json:"name"`
// }

// type Booking struct {
// 	ID      int       `json:"id"`
// 	RoomID  int       `json:"room_id"`
// 	UserID  int       `json:"user_id"`
// 	StartAt time.Time `json:"start_at"`
// 	EndAt   time.Time `json:"end_at"`
// }

// type Claims struct {
// 	Username string `json:"username"`
// 	Role     string `json:"role"`
// 	jwt.StandardClaims
// }

// func main() {
// 	var err error
// 	db, err = sql.Open("postgres", "user=postgres password=yourpassword dbname=meeting_rooms sslmode=disable")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()

// 	r := mux.NewRouter()
// 	r.HandleFunc("/register", registerUser).Methods("POST")
// 	r.HandleFunc("/login", loginUser).Methods("POST")
// 	r.HandleFunc("/rooms", getRooms).Methods("GET")
// 	r.HandleFunc("/rooms", createRoom).Methods("POST")
// 	r.HandleFunc("/rooms/{id}", deleteRoom).Methods("DELETE")
// 	r.HandleFunc("/book", createBooking).Methods("POST")
// 	r.HandleFunc("/book/{id}", cancelBooking).Methods("DELETE")
// 	r.HandleFunc("/bookings", getBookings).Methods("GET")
// 	http.ListenAndServe(":8080", r)
// }

// func registerUser(w http.ResponseWriter, r *http.Request) {
// 	var user User
// 	json.NewDecoder(r.Body).Decode(&user)
// 	_, err := db.Exec("INSERT INTO users (username, password, role) VALUES ($1, $2, $3)", user.Username, user.Password, "user")
// 	if err != nil {
// 		http.Error(w, "Error registering user", http.StatusInternalServerError)
// 		return
// 	}
// 	w.Write([]byte("User registered successfully"))
// }

// func loginUser(w http.ResponseWriter, r *http.Request) {
// 	var user User
// 	json.NewDecoder(r.Body).Decode(&user)
// 	var id int
// 	var role string
// 	err := db.QueryRow("SELECT id, role FROM users WHERE username=$1 AND password=$2", user.Username, user.Password).Scan(&id, &role)
// 	if err != nil {
// 		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
// 		return
// 	}
// 	expirationTime := time.Now().Add(24 * time.Hour)
// 	claims := &Claims{
// 		Username: user.Username,
// 		Role:     role,
// 		StandardClaims: jwt.StandardClaims{
// 			ExpiresAt: expirationTime.Unix(),
// 		},
// 	}
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tokenString, err := token.SignedString(jwtKey)
// 	if err != nil {
// 		http.Error(w, "Error generating token", http.StatusInternalServerError)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
// }

// func getRooms(w http.ResponseWriter, r *http.Request) {
// 	rows, err := db.Query("SELECT id, name FROM rooms")
// 	if err != nil {
// 		http.Error(w, "Error fetching rooms", http.StatusInternalServerError)
// 		return
// 	}
// 	defer rows.Close()
// 	var rooms []Room
// 	for rows.Next() {
// 		var room Room
// 		rows.Scan(&room.ID, &room.Name)
// 		rooms = append(rooms, room)
// 	}
// 	json.NewEncoder(w).Encode(rooms)
// }

// func createRoom(w http.ResponseWriter, r *http.Request) {
// 	var room Room
// 	json.NewDecoder(r.Body).Decode(&room)
// 	_, err := db.Exec("INSERT INTO rooms (name) VALUES ($1)", room.Name)
// 	if err != nil {
// 		http.Error(w, "Error creating room", http.StatusInternalServerError)
// 		return
// 	}
// 	w.Write([]byte("Room created successfully"))
// }

// func deleteRoom(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	_, err := db.Exec("DELETE FROM rooms WHERE id=$1", vars["id"])
// 	if err != nil {
// 		http.Error(w, "Error deleting room", http.StatusInternalServerError)
// 		return
// 	}
// 	w.Write([]byte("Room deleted successfully"))
// }

// func createBooking(w http.ResponseWriter, r *http.Request) {
// 	var booking Booking
// 	json.NewDecoder(r.Body).Decode(&booking)
// 	_, err := db.Exec("INSERT INTO bookings (room_id, user_id, start_at, end_at) VALUES ($1, $2, $3, $4)", booking.RoomID, booking.UserID, booking.StartAt, booking.EndAt)
// 	if err != nil {
// 		http.Error(w, "Error creating booking", http.StatusInternalServerError)
// 		return
// 	}
// 	w.Write([]byte("Booking created successfully"))
// }

// func cancelBooking(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	_, err := db.Exec("DELETE FROM bookings WHERE id=$1", vars["id"])
// 	if err != nil {
// 		http.Error(w, "Error cancelling booking", http.StatusInternalServerError)
// 		return
// 	}
// 	w.Write([]byte("Booking cancelled successfully"))
// }

// func getBookings(w http.ResponseWriter, r *http.Request) {
// 	rows, err := db.Query("SELECT id, room_id, user_id, start_at, end_at FROM bookings")
// 	if err != nil {
// 		http.Error(w, "Error fetching bookings", http.StatusInternalServerError)
// 		return
// 	}
// 	defer rows.Close()
// 	var bookings []Booking
// 	for rows.Next() {
// 		var booking Booking
// 		rows.Scan(&booking.ID, &booking.RoomID, &booking.UserID, &booking.StartAt, &booking.EndAt)
// 		bookings = append(bookings, booking)
// 	}
// 	json.NewEncoder(w).Encode(bookings)
// }
