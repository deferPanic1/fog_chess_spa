package models

import "time"

type Lobby struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`
	HostID        int64     `json:"hostId"`
	HostUsername  string    `json:"hostUsername"`
	GuestID       *int64    `json:"guestId,omitempty"`
	GuestUsername *string   `json:"guestUsername,omitempty"`
	Status        string    `json:"status"`
	TimeControl   int       `json:"timeControl"`
	HostReady     bool      `json:"hostReady"`
	GuestReady    bool      `json:"guestReady"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
