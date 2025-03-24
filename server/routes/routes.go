package routes

import (
	"github.com/sailesh-kona/meeting-room-booking-system/server/handler"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(r *mux.Router) {
	r.HandleFunc("/login", handler.Login).Methods("POST")
}

func RegisterRoomRoutes(r *mux.Router) {
	r.HandleFunc("/rooms", handler.GetRooms).Methods("GET")
	r.HandleFunc("/rooms", handler.CreateRoom).Methods("POST")
}

func RegisterBookingRoutes(r *mux.Router) {
	r.HandleFunc("/bookings", handler.GetBookings).Methods("GET")
	r.HandleFunc("/bookings", handler.CreateBooking).Methods("POST")
}
