package models

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	PassHash  []byte    `json:"-"`
	Rating    int32     `json:"rating"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserStats struct {
	Games  int `json:"games"`
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`
}

type MatchHistoryItem struct {
	Date         time.Time `json:"date"`
	Opponent     string    `json:"opponent"`
	Result       string    `json:"result"`
	RatingChange string    `json:"ratingChange"`
}

type AdminStats struct {
	UsersCount         int64 `json:"usersCount"`
	AdminsCount        int64 `json:"adminsCount"`
	MatchesCount       int64 `json:"matchesCount"`
	FinishedMatches    int64 `json:"finishedMatches"`
	OpenLobbiesCount   int64 `json:"openLobbiesCount"`
	ClosedLobbiesCount int64 `json:"closedLobbiesCount"`
}
