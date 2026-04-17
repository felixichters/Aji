package net

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	"nhooyr.io/websocket"

	"github.com/felixichters/Aji/server/internal/board"
	"github.com/felixichters/Aji/server/internal/game"
	"github.com/felixichters/Aji/server/internal/player"
	"github.com/felixichters/Aji/server/internal/protocol"
	"github.com/felixichters/Aji/server/internal/world"
)

// client is one connected WebSocket peer.
type client struct {
	conn     *websocket.Conn
	playerID player.ID // empty until joined
	send     chan []byte
}

// Hub manages all WebSocket connections for a single World.
type Hub struct {
	world  *world.World
	logger *log.Logger

	mu      sync.Mutex
	clients map[*client]struct{}
}

// New creates a Hub backed by the given World.
func New(w *world.World, logger *log.Logger) *Hub {
	return &Hub{
		world:   w,
		logger:  logger,
		clients: make(map[*client]struct{}),
	}
}

// ServeHTTP upgrades the connection to a WebSocket and runs the client
// loop. It satisfies http.Handler.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Vite dev proxy uses a different origin.
	})
	if err != nil {
		h.logger.Printf("ws accept: %v", err)
		return
	}

	c := &client{
		conn: conn,
		send: make(chan []byte, 16),
	}
	h.addClient(c)
	defer h.removeClient(c)

	ctx := r.Context()
	go h.writePump(ctx, c)
	h.readPump(ctx, c)
}

// readPump receives messages from one client and dispatches them.
func (h *Hub) readPump(ctx context.Context, c *client) {
	defer c.conn.Close(websocket.StatusNormalClosure, "")
	for {
		_, raw, err := c.conn.Read(ctx)
		if err != nil {
			if !isNormalClose(err) {
				h.logger.Printf("ws read: %v", err)
			}
			return
		}
		env, err := protocol.Decode(raw)
		if err != nil {
			h.sendError(c, "bad_request", "malformed message")
			continue
		}
		switch env.Type {
		case "join":
			h.handleJoin(c, env.Payload)
		case "place":
			h.handlePlace(c, env.Payload)
		default:
			h.sendError(c, "bad_request", "unknown message type: "+env.Type)
		}
	}
}

// writePump drains the send channel onto the socket.
func (h *Hub) writePump(ctx context.Context, c *client) {
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.conn.Write(ctx, websocket.MessageText, msg); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (h *Hub) handleJoin(c *client, raw json.RawMessage) {
	var msg protocol.JoinMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		h.sendError(c, "bad_request", "invalid join payload")
		return
	}
	pid := player.ID(msg.PlayerID)
	if _, err := h.world.Join(pid); err != nil {
		h.sendError(c, errorCode(err), err.Error())
		return
	}
	c.playerID = pid

	bw, bh := h.world.BoardSize()
	snap := h.world.Snapshot()
	joined, _ := protocol.Encode("joined", protocol.JoinedMsg{
		PlayerID: msg.PlayerID,
		BoardW:   bw,
		BoardH:   bh,
		Radius:   h.world.Radius(),
		State:    toStateSnapshot(snap),
	})
	c.send <- joined

	// Broadcast updated state so all clients see the new player.
	h.broadcast(snap)
}

func (h *Hub) handlePlace(c *client, raw json.RawMessage) {
	if c.playerID == "" {
		h.sendError(c, "not_joined", "must join before placing")
		return
	}
	var msg protocol.PlaceMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		h.sendError(c, "bad_request", "invalid place payload")
		return
	}
	if err := h.world.PlaceStone(c.playerID, board.Cell{X: msg.X, Y: msg.Y}); err != nil {
		h.sendError(c, errorCode(err), err.Error())
		return
	}
	h.broadcast(h.world.Snapshot())
}

func (h *Hub) broadcast(snap world.Snapshot) {
	msg, _ := protocol.Encode("state", toStateSnapshot(snap))
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		select {
		case c.send <- msg:
		default:
			// Client too slow — drop. The next broadcast carries full state.
		}
	}
}

func (h *Hub) sendError(c *client, code, message string) {
	msg, _ := protocol.Encode("error", protocol.ErrorMsg{Code: code, Message: message})
	select {
	case c.send <- msg:
	default:
	}
}

func (h *Hub) addClient(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

func (h *Hub) removeClient(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
	close(c.send)
}

// --- conversion helpers ---

func toStateSnapshot(s world.Snapshot) protocol.StateSnapshot {
	players := make([]protocol.PlayerState, len(s.Players))
	for i, p := range s.Players {
		stones := make([]protocol.Cell, len(p.Stones))
		for j, st := range p.Stones {
			stones[j] = protocol.Cell{X: st.X, Y: st.Y}
		}
		players[i] = protocol.PlayerState{
			ID:      string(p.ID),
			JoinSeq: p.JoinSeq,
			Stones:  stones,
		}
	}

	cliques := make([]protocol.CliqueState, len(s.Cliques))
	for i, cl := range s.Cliques {
		members := make([]string, len(cl.Members))
		for j, m := range cl.Members {
			members[j] = string(m)
		}
		cliques[i] = protocol.CliqueState{
			Members: members,
			ToMove:  string(cl.ToMove),
		}
	}

	engagements := make([]protocol.EngagementEdge, len(s.Engagements))
	for i, e := range s.Engagements {
		engagements[i] = protocol.EngagementEdge{A: string(e.A), B: string(e.B)}
	}

	return protocol.StateSnapshot{
		Players:     players,
		Cliques:     cliques,
		Engagements: engagements,
	}
}

// errorCode maps game errors to stable wire codes.
func errorCode(err error) string {
	switch {
	case errors.Is(err, game.ErrUnknownPlayer):
		return "unknown_player"
	case errors.Is(err, game.ErrDuplicatePlayer):
		return "duplicate_player"
	case errors.Is(err, game.ErrNotYourTurn):
		return "not_your_turn"
	case errors.Is(err, game.ErrOccupied):
		return "occupied"
	case errors.Is(err, game.ErrOutOfBounds):
		return "out_of_bounds"
	case errors.Is(err, game.ErrNotEngaged):
		return "not_engaged"
	case errors.Is(err, game.ErrOutsideRegion):
		return "outside_region"
	case errors.Is(err, game.ErrBootstrapMustEngage):
		return "bootstrap_must_engage"
	default:
		return "internal"
	}
}

func isNormalClose(err error) bool {
	return websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway
}
