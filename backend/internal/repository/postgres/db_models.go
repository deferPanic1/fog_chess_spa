package postgres

import "time"

type userModel struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Email     string    `gorm:"column:email"`
	Username  string    `gorm:"column:username"`
	PassHash  []byte    `gorm:"column:pass_hash"`
	Rating    int32     `gorm:"column:rating"`
	Role      string    `gorm:"column:role"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (userModel) TableName() string {
	return "users"
}

type lobbyModel struct {
	ID          string    `gorm:"column:id;primaryKey;default:gen_random_uuid()"`
	Code        string    `gorm:"column:code"`
	HostID      int64     `gorm:"column:host_id"`
	GuestID     *int64    `gorm:"column:guest_id"`
	HostReady   bool      `gorm:"column:host_ready"`
	GuestReady  bool      `gorm:"column:guest_ready"`
	Status      string    `gorm:"column:status"`
	TimeControl int       `gorm:"column:time_control"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (lobbyModel) TableName() string {
	return "lobbies"
}

type matchModel struct {
	ID               string     `gorm:"column:id;primaryKey;default:gen_random_uuid()"`
	WhitePlayer      int64      `gorm:"column:white_player"`
	BlackPlayer      int64      `gorm:"column:black_player"`
	Status           string     `gorm:"column:status"`
	Result           *string    `gorm:"column:result"`
	BoardFEN         string     `gorm:"column:board_fen"`
	CurrentTurn      string     `gorm:"column:current_turn"`
	WhiteRatingDelta int        `gorm:"column:white_rating_delta"`
	BlackRatingDelta int        `gorm:"column:black_rating_delta"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	FinishedAt       *time.Time `gorm:"column:finished_at"`
}

func (matchModel) TableName() string {
	return "matches"
}
