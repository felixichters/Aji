package protocol

import "encoding/json"

// Envelope is the outer wire frame for every message. Type is the
// discriminator; Payload carries the type-specific body.
type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Encode wraps a typed payload into an Envelope and marshals to JSON.
func Encode(msgType string, payload any) ([]byte, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{Type: msgType, Payload: p})
}

// Decode reads an Envelope from raw bytes.
func Decode(raw []byte) (Envelope, error) {
	var env Envelope
	err := json.Unmarshal(raw, &env)
	return env, err
}

// --- Client → Server ---

// JoinMsg is sent by a client that wants to join the game.
type JoinMsg struct {
	PlayerID string `json:"playerId"`
}

// PlaceMsg is sent when a joined client places a stone.
type PlaceMsg struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// --- Server → Client ---

// JoinedMsg is the response to a successful join.
type JoinedMsg struct {
	PlayerID string        `json:"playerId"`
	BoardW   int           `json:"boardW"`
	BoardH   int           `json:"boardH"`
	Radius   int           `json:"radius"`
	State    StateSnapshot `json:"state"`
}

// ErrorMsg is sent to a single client when their action fails.
type ErrorMsg struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// --- Snapshot (also used as the "state" broadcast message) ---

// StateSnapshot captures the full game state the client needs to render.
type StateSnapshot struct {
	Players     []PlayerState    `json:"players"`
	Cliques     []CliqueState    `json:"cliques"`
	Engagements []EngagementEdge `json:"engagements"`
}

// PlayerState is the wire representation of a single player.
type PlayerState struct {
	ID      string   `json:"id"`
	JoinSeq int      `json:"joinSeq"`
	Stones  []Cell   `json:"stones"`
}

// Cell is the JSON representation of a board coordinate.
type Cell struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// CliqueState represents one maximal clique and its turn pointer.
type CliqueState struct {
	Members []string `json:"members"`
	ToMove  string   `json:"toMove"`
}

// EngagementEdge is a pair of engaged player IDs.
type EngagementEdge struct {
	A string `json:"a"`
	B string `json:"b"`
}
