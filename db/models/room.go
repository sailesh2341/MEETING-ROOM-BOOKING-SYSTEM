package models

import "time"

type Room struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Capacity  int       `json:"capacity"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
}
