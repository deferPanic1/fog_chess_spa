package dto

import (
	"time"

	"github.com/Gilf4/fog_chess/internal/domain/models"
)

type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Rating    int32     `json:"rating"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewUserResponse(u *models.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Rating:    u.Rating,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

type StatsResponse struct {
	Games  int `json:"games"`
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`
}

type MatchHistoryItemResponse struct {
	Date         time.Time `json:"date"`
	Opponent     string    `json:"opponent"`
	Result       string    `json:"result"`
	RatingChange int       `json:"ratingChange"`
}
