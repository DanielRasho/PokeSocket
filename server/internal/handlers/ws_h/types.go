package ws_h

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/DanielRasho/PokeSocket/internal/services/matchmaking_s"
	"github.com/DanielRasho/PokeSocket/internal/services/users_s"
	"github.com/DanielRasho/PokeSocket/utils"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Message struct {
	Type    int             `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type PlayerID = pgtype.UUID

type Handler struct {
	DBClient           *pgxpool.Pool
	Validator          validator.Validate
	Connections        map[PlayerID]*Connection
	mu                 sync.RWMutex // Protect concurrent access to Connections map
	UserService        *users_s.UserService
	MatchmakingService *matchmaking_s.MatchmakingService
}

func NewHandler(
	dbClient *pgxpool.Pool,
	validator *validator.Validate,
	userService *users_s.UserService,
	matchmakingService *matchmaking_s.MatchmakingService) http.HandlerFunc {
	h := Handler{
		DBClient:           dbClient,
		Validator:          *validator,
		Connections:        make(map[pgtype.UUID]*Connection),
		UserService:        userService,
		MatchmakingService: matchmakingService,
	}
	return h.HandleRequest
}

// AddConnection safely adds a connection to the map
func (h *Handler) AddConnection(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Connections[conn.PlayerID] = conn
	log.Debug().
		Str("player_id", conn.PlayerID.String()).
		Str("username", conn.Username).
		Msg("Connection added")
}

// RemoveConnection safely removes and cleans up a connection
func (h *Handler) RemoveConnection(playerID PlayerID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conn, exists := h.Connections[playerID]; exists {
		// Remove from matchmaking queue if they were waiting
		h.MatchmakingService.RemoveFromQueue(playerID)

		conn.Close() // Properly close the connection
		delete(h.Connections, playerID)
		log.Debug().
			Str("player_id", playerID.String()).
			Str("username", conn.Username).
			Msg("Connection removed")
	}
}

// SendToPlayer sends a message via the buffered channel
func (h *Handler) SendToPlayer(playerID PlayerID, message Message) error {
	h.mu.RLock()
	conn, exists := h.Connections[playerID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("player not connected")
	}

	select {
	case conn.Send <- message:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending message")
	default:
		log.Warn().Str("player_id", playerID.String()).Msg("Send channel full, message dropped")
		return fmt.Errorf("send channel full")
	}
}

func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Err(err).Msg("Failed to accept websocket")
		return
	}
	ctx := context.Background()

	// Setup cleanup on disconnect
	defer func() {
		log.Debug().Msg("Connection closing, cleaning up...")
		conn.Close(websocket.StatusNormalClosure, "Goodbye")
	}()

	// Read initial connect message
	var initialMsg Message
	err = wsjson.Read(ctx, conn, &initialMsg)
	if err != nil {
		sendAndLogError(ctx, conn, fmt.Errorf("Failed to read initial message"), initialMsg, utils.InvalidFields,
			map[string]string{"type": "Invalid request fields"})
		return
	}

	// Register connection
	connection, err := h.handleConnect(initialMsg, ctx, conn)
	if err != nil {
		if verr, ok := err.(*utils.VerificationError); ok {
			sendAndLogError(ctx, conn, err, initialMsg, utils.InvalidFields, verr.UserError)
		} else {
			sendAndLogError(ctx, conn, err, initialMsg, utils.InvalidFields,
				map[string]string{"type": "Error creating user"})
		}
		return
	}

	h.AddConnection(connection)
	defer h.RemoveConnection(connection.PlayerID)

	// Start background goroutines
	go connection.writePump()
	go connection.pingConnection()

	// HANDLE OTHER MESSAGES (blocks until disconnect)
	h.processMessages(connection)
}

func (h *Handler) processMessages(conn *Connection) {
	defer func() {
		log.Info().Str("username", conn.Username).Msg("Message processing stopped")
	}()

	for {
		var msg Message
		err := wsjson.Read(conn.Ctx, conn.Conn, &msg)
		if err != nil {
			// Check if it's a normal closure
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				log.Info().Str("username", conn.Username).Msg("User closed connection normally")
			} else if conn.Ctx.Err() != nil {
				log.Debug().Str("username", conn.Username).Msg("Context cancelled")
			} else {
				log.Err(err).Str("username", conn.Username).Msg("Error reading message")
			}
			return
		}

		// HANDLE MESSAGE - now using the channel instead of direct writes
		switch msg.Type {
		case CLIENT_MESSAGE_TYPE.Connect:
			conn.Send <- NewMessage(SERVER_MESSAGE_TYPE.Error, ErrorResponse{
				Message: "Already connected",
				Code:    utils.InvalidFields.StatusCode,
				Details: map[string]string{"type": "Already connected"},
			})

		case CLIENT_MESSAGE_TYPE.Status:
			conn.Send <- NewMessage(SERVER_MESSAGE_TYPE.Status, map[string]string{
				"status":   "connected",
				"username": conn.Username,
			})

		case CLIENT_MESSAGE_TYPE.Match:
			log.Debug().Str("username", conn.Username).Msg("Match request received")
			h.handleMatch(conn)

		case CLIENT_MESSAGE_TYPE.Attack:
			log.Debug().Str("username", conn.Username).Msg("Attack received")
			// TODO: Handle attack

		case CLIENT_MESSAGE_TYPE.ChangePokemon:
			log.Debug().Str("username", conn.Username).Msg("Change Pokemon received")
			// TODO: Handle pokemon change

		case CLIENT_MESSAGE_TYPE.Surrender:
			log.Debug().Str("username", conn.Username).Msg("Surrender received")
			// TODO: Handle surrender

		default:
			conn.Send <- NewMessage(SERVER_MESSAGE_TYPE.Error, ErrorResponse{
				Message: "Unknown message type",
				Code:    utils.InvalidFields.StatusCode,
				Details: map[string]string{"received_type": fmt.Sprint(msg.Type)},
			})
		}
	}
}
