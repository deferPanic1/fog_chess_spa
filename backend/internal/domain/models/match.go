package models

import "time"

type Match struct {
	ID               string     `json:"id"`
	WhitePlayer      int64      `json:"whitePlayer"`
	BlackPlayer      int64      `json:"blackPlayer"`
	Status           string     `json:"status"`
	Result           *string    `json:"result,omitempty"`
	BoardFEN         string     `json:"boardFen"`
	CurrentTurn      string     `json:"currentTurn"`
	WhiteRatingDelta int       `json:"whiteRatingDelta"`
	BlackRatingDelta int       `json:"blackRatingDelta"`
	CreatedAt        time.Time  `json:"createdAt"`
	FinishedAt       *time.Time `json:"finishedAt,omitempty"`
}
