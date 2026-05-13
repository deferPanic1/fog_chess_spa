package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	chessengine "github.com/Gilf4/fog_chess/internal/chessEngine"
	"github.com/Gilf4/fog_chess/internal/domain/models"
	"github.com/Gilf4/fog_chess/internal/repository"
)

var (
	ErrMatchAccessDenied = errors.New("user is not a match participant")
	ErrMatchNotActive    = errors.New("match is not active")
	ErrNotPlayersTurn    = errors.New("it is not the player's turn")
)

const eloKFactor = 32

type MatchSnapshot struct {
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
}

type ActiveMatch struct {
	MatchID          string    `json:"matchId"`
	Status           string    `json:"status"`
	YourColor        string    `json:"yourColor"`
	OpponentID       int64     `json:"opponentId"`
	OpponentUsername string    `json:"opponentUsername"`
	CreatedAt        time.Time `json:"createdAt"`
}

type MatchMoveResult struct {
	LastMove []string
	Notation string
	Color    string
	Finished bool
}

type MatchService struct {
	matches repository.MatchRepository
	users   repository.UserRepository

	mu           sync.RWMutex
	timeControls map[string]int
}

func NewMatchService(matches repository.MatchRepository, users repository.UserRepository) *MatchService {
	return &MatchService{
		matches:      matches,
		users:        users,
		timeControls: make(map[string]int),
	}
}

func (s *MatchService) StartFromLobby(ctx context.Context, lobby *models.Lobby) (*models.Match, error) {
	if lobby == nil || lobby.GuestID == nil {
		return nil, fmt.Errorf("lobby is incomplete")
	}

	if active, err := s.matches.GetActiveByUser(ctx, lobby.HostID); err == nil {
		if samePlayers(active, lobby.HostID, *lobby.GuestID) {
			return active, nil
		}
		return nil, repository.ErrUserAlreadyInMatch
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	if active, err := s.matches.GetActiveByUser(ctx, *lobby.GuestID); err == nil {
		if samePlayers(active, lobby.HostID, *lobby.GuestID) {
			return active, nil
		}
		return nil, repository.ErrUserAlreadyInMatch
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	match, err := s.matches.Create(ctx, lobby.HostID, *lobby.GuestID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.timeControls[match.ID] = lobby.TimeControl
	s.mu.Unlock()

	return match, nil
}

func (s *MatchService) GetByID(ctx context.Context, matchID string) (*models.Match, error) {
	return s.matches.GetByID(ctx, matchID)
}

func (s *MatchService) GetActiveByUser(ctx context.Context, userID int64) (*ActiveMatch, error) {
	match, err := s.matches.GetActiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	playerColor, err := colorForPlayer(match, userID)
	if err != nil {
		return nil, err
	}

	opponentID := match.WhitePlayer
	if opponentID == userID {
		opponentID = match.BlackPlayer
	}

	opponent, err := s.users.GetByID(ctx, opponentID)
	if err != nil {
		return nil, err
	}

	return &ActiveMatch{
		MatchID:          match.ID,
		Status:           match.Status,
		YourColor:        chessengine.ColorName(playerColor),
		OpponentID:       opponentID,
		OpponentUsername: opponent.Username,
		CreatedAt:        match.CreatedAt,
	}, nil
}

func (s *MatchService) GetSnapshot(ctx context.Context, matchID string, userID int64) (*MatchSnapshot, error) {
	match, err := s.matches.GetByID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	return s.buildSnapshot(ctx, match, userID, nil)
}

func (s *MatchService) ApplyMove(ctx context.Context, matchID string, userID int64, from, to, promotion string) (*MatchMoveResult, error) {
	match, err := s.matches.GetByID(ctx, matchID)
	if err != nil {
		return nil, err
	}
	if match.Status != "active" {
		return nil, ErrMatchNotActive
	}

	playerColor, err := colorForPlayer(match, userID)
	if err != nil {
		return nil, err
	}

	game, err := chessengine.NewGameFromFEN(match.BoardFEN)
	if err != nil {
		return nil, fmt.Errorf("parse match board: %w", err)
	}
	if game.Board.SideToMove != playerColor {
		return nil, ErrNotPlayersTurn
	}

	move, err := chessengine.ParseMove(from, to, promotion)
	if err != nil {
		return nil, err
	}
	move = defaultPromotionMove(game.Board, move)

	playerView := chessengine.GameState{Board: game.Board.BoardForPlayer(playerColor)}
	if _, err := playerView.DoMove(move); err != nil {
		return nil, err
	}

	finished, err := game.DoMove(move)
	if err != nil {
		return nil, err
	}

	updatedFEN, err := game.FEN()
	if err != nil {
		return nil, fmt.Errorf("format updated board: %w", err)
	}

	match.BoardFEN = updatedFEN
	match.CurrentTurn = chessengine.ColorName(game.Board.SideToMove)
	if finished {
		match.Status = "finished"
		result := winnerResult(playerColor)
		match.Result = &result
		now := time.Now().UTC()
		match.FinishedAt = &now
	}

	if finished {
		if err := s.finalizeMatchRatings(ctx, match); err != nil {
			return nil, err
		}
	} else {
		if err := s.matches.Update(ctx, match); err != nil {
			return nil, err
		}
	}

	return &MatchMoveResult{
		LastMove: []string{strings.ToLower(strings.TrimSpace(from)), strings.ToLower(strings.TrimSpace(to))},
		Notation: moveNotation(from, to, promotion),
		Color:    chessengine.ColorName(playerColor),
		Finished: finished,
	}, nil
}

func defaultPromotionMove(board chessengine.Board, move chessengine.Move) chessengine.Move {
	if move.Promotion != chessengine.Empty {
		return move
	}

	piece := board.Squares[move.From]
	if piece.Type != chessengine.Pawn {
		return move
	}

	toRank := move.To / 8
	if (piece.Color == chessengine.White && toRank == 0) || (piece.Color == chessengine.Black && toRank == 7) {
		move.Promotion = chessengine.Queen
	}

	return move
}

func (s *MatchService) Resign(ctx context.Context, matchID string, userID int64) error {
	match, err := s.matches.GetByID(ctx, matchID)
	if err != nil {
		return err
	}
	if match.Status != "active" {
		return ErrMatchNotActive
	}

	playerColor, err := colorForPlayer(match, userID)
	if err != nil {
		return err
	}

	match.Status = "finished"
	result := winnerResult(oppositeColor(playerColor))
	match.Result = &result
	now := time.Now().UTC()
	match.FinishedAt = &now

	return s.finalizeMatchRatings(ctx, match)
}

func (s *MatchService) buildSnapshot(ctx context.Context, match *models.Match, userID int64, lastMove []string) (*MatchSnapshot, error) {
	playerColor, err := colorForPlayer(match, userID)
	if err != nil {
		return nil, err
	}

	player, opponent, err := s.loadPlayers(ctx, match, userID)
	if err != nil {
		return nil, err
	}

	board, err := chessengine.ParseFEN(match.BoardFEN)
	if err != nil {
		return nil, fmt.Errorf("parse board fen: %w", err)
	}

	visibleFEN, err := board.DisplayFENForPlayer(playerColor)
	if err != nil {
		return nil, fmt.Errorf("build player fen: %w", err)
	}

	playerBoard := board.BoardForPlayer(playerColor)
	legalMoves := map[string][]string{}
	if match.Status == "active" {
		legalMoves = playerBoard.LegalMovesFor(playerColor)
	}

	return &MatchSnapshot{
		Status:              match.Status,
		Result:              resultForPlayer(match.Result, playerColor),
		YourColor:           chessengine.ColorName(playerColor),
		PlayerUsername:      player.Username,
		OpponentUsername:    opponent.Username,
		PlayerRating:        player.Rating,
		OpponentRating:      opponent.Rating,
		PlayerRatingDelta:   ratingDeltaForPlayer(match, userID),
		OpponentRatingDelta: ratingDeltaForOpponent(match, userID),
		FEN:                 visibleFEN,
		TurnColor:           chessengine.ColorName(board.SideToMove),
		PlayerTime:          s.timeControlFor(match.ID),
		OpponentTime:        s.timeControlFor(match.ID),
		FoggedSquares:       board.FoggedSquaresFor(playerColor),
		LegalMoves:          legalMoves,
		LastMove:            lastMove,
	}, nil
}

func (s *MatchService) loadPlayers(ctx context.Context, match *models.Match, userID int64) (*models.User, *models.User, error) {
	player, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	opponentID := match.WhitePlayer
	if opponentID == userID {
		opponentID = match.BlackPlayer
	}

	opponent, err := s.users.GetByID(ctx, opponentID)
	if err != nil {
		return nil, nil, err
	}

	return player, opponent, nil
}

func (s *MatchService) timeControlFor(matchID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if seconds, ok := s.timeControls[matchID]; ok && seconds > 0 {
		return seconds
	}

	return 600
}

func (s *MatchService) finalizeMatchRatings(ctx context.Context, match *models.Match) error {
	white, err := s.users.GetByID(ctx, match.WhitePlayer)
	if err != nil {
		return err
	}

	black, err := s.users.GetByID(ctx, match.BlackPlayer)
	if err != nil {
		return err
	}

	whiteDelta, blackDelta := calculateEloDelta(white.Rating, black.Rating, match.Result)
	match.WhiteRatingDelta = whiteDelta
	match.BlackRatingDelta = blackDelta

	return s.matches.Complete(ctx, match, white.Rating+int32(whiteDelta), black.Rating+int32(blackDelta))
}

func calculateEloDelta(whiteRating, blackRating int32, result *string) (int, int) {
	whiteScore, _ := matchScores(result)
	whiteExpected := expectedScore(whiteRating, blackRating)

	whiteDelta := int(math.Round(float64(eloKFactor) * (whiteScore - whiteExpected)))
	blackDelta := -whiteDelta

	return whiteDelta, blackDelta
}

func expectedScore(playerRating, opponentRating int32) float64 {
	return 1 / (1 + math.Pow(10, float64(opponentRating-playerRating)/400))
}

func matchScores(result *string) (float64, float64) {
	if result == nil {
		return 0.5, 0.5
	}

	switch *result {
	case "white", "white_win":
		return 1, 0
	case "black", "black_win":
		return 0, 1
	case "draw":
		return 0.5, 0.5
	default:
		return 0.5, 0.5
	}
}

func ratingDeltaForPlayer(match *models.Match, userID int64) int {
	if userID == match.WhitePlayer {
		return match.WhiteRatingDelta
	}
	return match.BlackRatingDelta
}

func ratingDeltaForOpponent(match *models.Match, userID int64) int {
	if userID == match.WhitePlayer {
		return match.BlackRatingDelta
	}
	return match.WhiteRatingDelta
}

func colorForPlayer(match *models.Match, userID int64) (chessengine.Color, error) {
	switch userID {
	case match.WhitePlayer:
		return chessengine.White, nil
	case match.BlackPlayer:
		return chessengine.Black, nil
	default:
		return chessengine.NoColor, ErrMatchAccessDenied
	}
}

func oppositeColor(color chessengine.Color) chessengine.Color {
	if color == chessengine.White {
		return chessengine.Black
	}
	return chessengine.White
}

func winnerResult(color chessengine.Color) string {
	if color == chessengine.White {
		return "white_win"
	}
	return "black_win"
}

func resultForPlayer(result *string, color chessengine.Color) string {
	if result == nil {
		return ""
	}

	switch *result {
	case "draw":
		return "draw"
	case "white_win":
		if color == chessengine.White {
			return "win"
		}
		return "loss"
	case "black_win":
		if color == chessengine.Black {
			return "win"
		}
		return "loss"
	default:
		return *result
	}
}

func moveNotation(from, to, promotion string) string {
	from = strings.ToLower(strings.TrimSpace(from))
	to = strings.ToLower(strings.TrimSpace(to))
	if promotion == "" {
		return from + "-" + to
	}
	return fmt.Sprintf("%s-%s=%s", from, to, strings.ToLower(strings.TrimSpace(promotion)))
}

func samePlayers(match *models.Match, userA, userB int64) bool {
	if match == nil {
		return false
	}
	return (match.WhitePlayer == userA && match.BlackPlayer == userB) ||
		(match.WhitePlayer == userB && match.BlackPlayer == userA)
}
