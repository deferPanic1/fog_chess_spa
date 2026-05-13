package ws

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Gilf4/fog_chess/internal/repository"
)

func (h *Handler) handleReady(c *Client, payload json.RawMessage) {
	var p ReadyPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}

	lobby, err := h.lobbyService.SetReady(context.Background(), c.LobbyCode, c.UserID, p.IsReady)
	if err != nil {
		h.sendError(c, "failed to update ready state")
		return
	}

	readyRaw, _ := json.Marshal(ReadyBroadcast{PlayerID: c.UserID, IsReady: p.IsReady})
	h.broadcast(c.LobbyCode, Message{Type: "ready", Payload: readyRaw})

	statusRaw, _ := json.Marshal(LobbyStatusBroadcast{Status: lobby.Status})
	h.broadcast(c.LobbyCode, Message{Type: "lobby_status", Payload: statusRaw})
}

func (h *Handler) handleStart(c *Client) {
	lobby, err := h.lobbyService.GetByCode(context.Background(), c.LobbyCode)
	if err != nil {
		h.sendError(c, "lobby not found")
		return
	}

	if lobby.HostID != c.UserID {
		h.sendError(c, "only host can start the match")
		return
	}

	if lobby.Status != "ready" {
		h.sendError(c, "both players must be ready")
		return
	}

	match, err := h.matchService.StartFromLobby(context.Background(), lobby)
	if err != nil {
		h.sendError(c, startMatchErrorMessage(err))
		return
	}

	if err := h.lobbyService.Close(context.Background(), c.LobbyCode); err != nil {
		h.sendError(c, "failed to close lobby after start")
		return
	}

	matchRaw, _ := json.Marshal(MatchStartBroadcast{
		MatchID:     match.ID,
		WhitePlayer: match.WhitePlayer,
		BlackPlayer: match.BlackPlayer,
	})
	h.broadcast(c.LobbyCode, Message{Type: "match_start", Payload: matchRaw})
	h.broadcastLobbyState(c.LobbyCode)
}

func startMatchErrorMessage(err error) string {
	if errors.Is(err, repository.ErrUserAlreadyInMatch) {
		return "one of the players already has an active match"
	}
	return "failed to start match"
}

func (h *Handler) broadcastLobbyState(lobbyCode string) {
	lobby, err := h.lobbyService.GetByCode(context.Background(), lobbyCode)
	if err != nil {
		return
	}

	statusRaw, _ := json.Marshal(LobbyStatusBroadcast{Status: lobby.Status})
	h.broadcast(lobbyCode, Message{Type: "lobby_status", Payload: statusRaw})

	hostRaw, _ := json.Marshal(PlayerSyncBroadcast{
		PlayerID:  lobby.HostID,
		Role:      "host",
		Username:  lobby.HostUsername,
		IsReady:   lobby.HostReady,
		IsPresent: true,
	})
	h.broadcast(lobbyCode, Message{Type: "player_sync", Payload: hostRaw})

	guestRaw, _ := json.Marshal(PlayerSyncBroadcast{
		PlayerID:  int64Value(lobby.GuestID),
		Role:      "guest",
		Username:  stringValue(lobby.GuestUsername),
		IsReady:   lobby.GuestReady,
		IsPresent: lobby.GuestID != nil,
	})
	h.broadcast(lobbyCode, Message{Type: "player_sync", Payload: guestRaw})
}

func int64Value(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
