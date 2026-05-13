package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Gilf4/fog_chess/internal/repository"
	"github.com/Gilf4/fog_chess/internal/service"
	"github.com/Gilf4/fog_chess/internal/transport/http/middleware"
)

func (h *Handler) HandleMatch(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	matchID := strings.TrimSpace(r.URL.Query().Get("matchId"))
	if matchID == "" {
		http.Error(w, "missing matchId", http.StatusBadRequest)
		return
	}

	if _, err := h.matchService.GetSnapshot(r.Context(), matchID, userID); err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			http.Error(w, "match not found", http.StatusNotFound)
		case errors.Is(err, service.ErrMatchAccessDenied):
			http.Error(w, "forbidden", http.StatusForbidden)
		default:
			http.Error(w, "failed to load match", http.StatusInternalServerError)
		}
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("match ws upgrade failed", "err", err)
		return
	}

	client := &Client{
		UserID:  userID,
		MatchID: matchID,
		Conn:    conn,
		Send:    make(chan any, 8),
	}

	h.registerMatch(client)
	h.pushMatchSnapshot(client, "match_init", nil, nil)

	go h.writePump(client)
	h.readMatchPump(client)
}

func (h *Handler) registerMatch(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.matches[c.MatchID]; !ok {
		h.matches[c.MatchID] = make(map[int64]*Client)
	}
	h.matches[c.MatchID][c.UserID] = c
}

func (h *Handler) unregisterMatch(c *Client) {
	h.mu.Lock()
	if room, ok := h.matches[c.MatchID]; ok {
		delete(room, c.UserID)
		if len(room) == 0 {
			delete(h.matches, c.MatchID)
		}
	}
	close(c.Send)
	c.Conn.Close()
	h.mu.Unlock()
}

func (h *Handler) readMatchPump(c *Client) {
	defer h.unregisterMatch(c)

	c.Conn.SetReadLimit(2048)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg Message
		if err := c.Conn.ReadJSON(&msg); err != nil {
			break
		}

		switch msg.Type {
		case "move":
			h.handleMatchMove(c, msg.Payload)
		case "resign":
			h.handleResign(c)
		}
	}
}

func (h *Handler) handleMatchMove(c *Client, payload json.RawMessage) {
	var move MatchMovePayload
	if err := json.Unmarshal(payload, &move); err != nil {
		h.sendError(c, "invalid move payload")
		return
	}

	result, err := h.matchService.ApplyMove(context.Background(), c.MatchID, c.UserID, move.From, move.To, move.Promotion)
	if err != nil {
		h.sendMatchError(c, err)
		return
	}

	moveMeta := &MatchMoveMeta{
		SAN:   result.Notation,
		Color: result.Color,
	}
	h.broadcastMatchState(c.MatchID, result.LastMove, moveMeta)
}

func (h *Handler) handleResign(c *Client) {
	if err := h.matchService.Resign(context.Background(), c.MatchID, c.UserID); err != nil {
		h.sendMatchError(c, err)
		return
	}

	h.broadcastMatchState(c.MatchID, nil, nil)
}

func (h *Handler) broadcastMatchState(matchID string, lastMove []string, moveMeta *MatchMoveMeta) {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.matches[matchID]))
	for _, client := range h.matches[matchID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	messageType := "board_update"
	if moveMeta == nil && lastMove == nil {
		messageType = "game_over"
	}

	for _, client := range clients {
		h.pushMatchSnapshot(client, messageType, lastMove, moveMeta)
	}
}

func (h *Handler) pushMatchSnapshot(c *Client, messageType string, lastMove []string, moveMeta *MatchMoveMeta) {
	snapshot, err := h.matchService.GetSnapshot(context.Background(), c.MatchID, c.UserID)
	if err != nil {
		h.sendError(c, "failed to load match snapshot")
		return
	}

	payload := MatchSnapshotPayload{
		Status:              snapshot.Status,
		Result:              snapshot.Result,
		YourColor:           snapshot.YourColor,
		PlayerUsername:      snapshot.PlayerUsername,
		OpponentUsername:    snapshot.OpponentUsername,
		PlayerRating:        snapshot.PlayerRating,
		OpponentRating:      snapshot.OpponentRating,
		PlayerRatingDelta:   snapshot.PlayerRatingDelta,
		OpponentRatingDelta: snapshot.OpponentRatingDelta,
		FEN:                 snapshot.FEN,
		TurnColor:           snapshot.TurnColor,
		PlayerTime:          snapshot.PlayerTime,
		OpponentTime:        snapshot.OpponentTime,
		FoggedSquares:       snapshot.FoggedSquares,
		LegalMoves:          snapshot.LegalMoves,
		LastMove:            lastMove,
		Move:                moveMeta,
	}

	raw, _ := json.Marshal(payload)
	select {
	case c.Send <- Message{Type: messageType, Payload: raw}:
	default:
		go h.unregisterMatch(c)
	}
}

func (h *Handler) sendMatchError(c *Client, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		h.sendError(c, "match not found")
	case errors.Is(err, service.ErrMatchAccessDenied):
		h.sendError(c, "you are not a participant of this match")
	case errors.Is(err, service.ErrMatchNotActive):
		h.sendError(c, "match is already finished")
	case errors.Is(err, service.ErrNotPlayersTurn):
		h.sendError(c, "it is not your turn")
	default:
		h.sendError(c, err.Error())
	}
}
