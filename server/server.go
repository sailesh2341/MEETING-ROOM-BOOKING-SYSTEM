package server

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sailesh-kona/meeting-room-booking-system/config"
	"github.com/sailesh-kona/meeting-room-booking-system/server/handler"
	"github.com/sailesh-kona/meeting-room-booking-system/server/middleware"
)

func New(database *sql.DB, cfg config.Config) http.Handler {
	router := mux.NewRouter()
	api := handler.New(database, cfg)
	auth := middleware.Auth(cfg.JWTSecret)
	admin := chain(auth, middleware.RequireRole("admin"))

	router.HandleFunc("/health", api.Health).Methods(http.MethodGet)
	router.HandleFunc("/register", api.Register).Methods(http.MethodPost)
	router.HandleFunc("/login", api.Login).Methods(http.MethodPost)

	router.Handle("/rooms", auth(http.HandlerFunc(api.ListRooms))).Methods(http.MethodGet)
	router.Handle("/rooms/{id}", auth(http.HandlerFunc(api.GetRoom))).Methods(http.MethodGet)
	router.Handle("/bookings", auth(http.HandlerFunc(api.CreateBooking))).Methods(http.MethodPost)
	router.Handle("/bookings/me", auth(http.HandlerFunc(api.ListMyBookings))).Methods(http.MethodGet)
	router.Handle("/bookings/{id}", auth(http.HandlerFunc(api.UpdateBooking))).Methods(http.MethodPut)
	router.Handle("/bookings/{id}", auth(http.HandlerFunc(api.CancelBooking))).Methods(http.MethodDelete)

	router.Handle("/users", admin(http.HandlerFunc(api.ListUsers))).Methods(http.MethodGet)
	router.Handle("/users/{id}/role", admin(http.HandlerFunc(api.UpdateUserRole))).Methods(http.MethodPut)
	router.Handle("/users/{id}", admin(http.HandlerFunc(api.DeleteUser))).Methods(http.MethodDelete)
	router.Handle("/rooms", admin(http.HandlerFunc(api.CreateRoom))).Methods(http.MethodPost)
	router.Handle("/rooms/{id}", admin(http.HandlerFunc(api.UpdateRoom))).Methods(http.MethodPut)
	router.Handle("/rooms/{id}", admin(http.HandlerFunc(api.DeleteRoom))).Methods(http.MethodDelete)
	router.Handle("/bookings", admin(http.HandlerFunc(api.ListBookings))).Methods(http.MethodGet)
	router.Handle("/audit-logs", admin(http.HandlerFunc(api.ListAuditLogs))).Methods(http.MethodGet)

	return router
}

func chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
