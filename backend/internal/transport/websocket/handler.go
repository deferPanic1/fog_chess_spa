package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Gilf4/fog_chess/internal/domain/models"
	"github.com/Gilf4/fog_chess/internal/service"
	"github.com/Gilf4/fog_chess/internal/transport/http/middleware"
	"github.com/gorilla/websocket"
)

type lobbyService interface {
	GetByCode(ctx context.Context, code string) (*models.Lobby, error)
	SetReady(ctx context.Context, code string, userID int64, ready bool) (*models.Lobby, error)
	Close(ctx context.Context, code string) error
}

type matchService interface {
	StartFromLobby(ctx context.Context, lobby *models.Lobby) (*models.Match, error)
	GetByID(ctx context.Context, matchID string) (*models.Match, error)
	GetSnapshot(ctx context.Context, matchID string, userID int64) (*service.MatchSnapshot, error)
	ApplyMove(ctx context.Context, matchID string, userID int64, from, to, promotion string) (*service.MatchMoveResult, error)
	Resign(ctx context.Context, matchID string, userID int64) error
}

type Client struct {
	UserID    int64
	LobbyCode string
	MatchID   string
	Conn      *websocket.Conn
	Send      chan any
}

type Handler struct {
	log          *slog.Logger
	lobbyService lobbyService
	matchService matchService
	upgrader     websocket.Upgrader

	mu      sync.RWMutex
	lobbies map[string]map[int64]*Client
	matches map[string]map[int64]*Client
}

func New(log *slog.Logger, lobbySvc lobbyService, matchSvc matchService) *Handler {
	return &Handler{
		log:          log,
		lobbyService: lobbySvc,
		matchService: matchSvc,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(*http.Request) bool { return true },
		},
		lobbies: make(map[string]map[int64]*Client),
		matches: make(map[string]map[int64]*Client),
	}
}

func (h *Handler) HandleLobby(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	lobbyCode := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("lobbyCode")))
	if lobbyCode == "" {
		http.Error(w, "missing lobbyCode", http.StatusBadRequest)
		return
	}

	lobby, err := h.lobbyService.GetByCode(r.Context(), lobbyCode)
	if err != nil {
		http.Error(w, "lobby not found", http.StatusNotFound)
		return
	}

	if lobby.HostID != userID && (lobby.GuestID == nil || *lobby.GuestID != userID) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if !h.canConnect(lobbyCode, userID) {
		http.Error(w, "lobby is full", http.StatusForbidden)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("ws upgrade failed", slog.Any("err", err))
		return
	}

	client := &Client{
		UserID:    userID,
		LobbyCode: lobbyCode,
		Conn:      conn,
		Send:      make(chan any, 8),
	}

	h.registerLobby(client)
	h.broadcastLobbyState(lobbyCode)

	go h.writePump(client)
	h.readLobbyPump(client)
}

func (h *Handler) canConnect(lobbyCode string, userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.lobbies[lobbyCode]
	if !ok {
		return true
	}
	if _, exists := clients[userID]; exists {
		return true
	}
	return len(clients) < 2
}

func (h *Handler) registerLobby(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.lobbies[c.LobbyCode]; !ok {
		h.lobbies[c.LobbyCode] = make(map[int64]*Client)
	}
	h.lobbies[c.LobbyCode][c.UserID] = c
}

func (h *Handler) unregisterLobby(c *Client) {
	h.mu.Lock()
	if room, ok := h.lobbies[c.LobbyCode]; ok {
		delete(room, c.UserID)
		if len(room) == 0 {
			delete(h.lobbies, c.LobbyCode)
		}
	}
	close(c.Send)
	c.Conn.Close()
	h.mu.Unlock()

	h.broadcastLobbyState(c.LobbyCode)
}

func (h *Handler) readLobbyPump(c *Client) {
	defer h.unregisterLobby(c)

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
		case "ready":
			h.handleReady(c, msg.Payload)
		case "start":
			h.handleStart(c)
		}
	}
}

func (h *Handler) writePump(c *Client) {
	ticker := time.NewTicker(50 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Handler) broadcast(lobbyCode string, msg any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.lobbies[lobbyCode] {
		select {
		case c.Send <- msg:
		default:
			go h.unregisterLobby(c)
		}
	}
}

func (h *Handler) sendError(c *Client, message string) {
	raw, _ := json.Marshal(ErrorBroadcast{Message: message})
	select {
	case c.Send <- Message{Type: "error", Payload: raw}:
	default:
		if c.LobbyCode != "" {
			go h.unregisterLobby(c)
			return
		}
		go h.unregisterMatch(c)
	}
}
