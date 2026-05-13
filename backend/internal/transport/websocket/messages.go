package ws

import "encoding/json"

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type ReadyPayload struct {
	IsReady bool `json:"isReady"`
}

type ReadyBroadcast struct {
	PlayerID int64 `json:"playerId"`
	IsReady  bool  `json:"isReady"`
}

type LobbyStatusBroadcast struct {
	Status string `json:"status"`
}

type PlayerSyncBroadcast struct {
	PlayerID  int64  `json:"playerId"`
	Role      string `json:"role"`
	Username  string `json:"username"`
	IsReady   bool   `json:"isReady"`
	IsPresent bool   `json:"isPresent"`
}

type MatchStartBroadcast struct {
	MatchID     string `json:"matchId"`
	WhitePlayer int64  `json:"whitePlayer"`
	BlackPlayer int64  `json:"blackPlayer"`
}

type ErrorBroadcast struct {
	Message string `json:"message"`
}

type MatchMovePayload struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Promotion string `json:"promotion,omitempty"`
}

type MatchMoveMeta struct {
	SAN   string `json:"san"`
	Color string `json:"color"`
}

type MatchSnapshotPayload struct {
	Status           string              `json:"status"`
	Result           string              `json:"result,omitempty"`
	YourColor        string              `json:"yourColor"`
	PlayerUsername   string              `json:"playerUsername"`
	OpponentUsername string              `json:"opponentUsername"`
	PlayerRating     int32               `json:"playerRating"`
	OpponentRating   int32               `json:"opponentRating"`
	PlayerRatingDelta int                `json:"playerRatingDelta"`
	OpponentRatingDelta int              `json:"opponentRatingDelta"`
	FEN              string              `json:"fen"`
	TurnColor        string              `json:"turnColor"`
	PlayerTime       int                 `json:"playerTime"`
	OpponentTime     int                 `json:"opponentTime"`
	FoggedSquares    []string            `json:"foggedSquares"`
	LegalMoves       map[string][]string `json:"legalMoves"`
	LastMove         []string            `json:"lastMove,omitempty"`
	Move             *MatchMoveMeta      `json:"move,omitempty"`
}
