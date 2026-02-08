package ws_h

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DanielRasho/PokeSocket/utils"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Message struct {
	Type    int             `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type User struct {
	Username string    `json:"username"`
	Id       uuid.UUID `json:"uuid"`
}

type Handler struct {
	DBClient  *pgxpool.Pool
	Validator validator.Validate
	Sessions  map[uuid.UUID]User
}

func NewHandler(dbClient *pgxpool.Pool, validator *validator.Validate) http.HandlerFunc {
	h := Handler{
		DBClient:  dbClient,
		Validator: *validator,
		Sessions:  make(map[uuid.UUID]User),
	}
	return h.HandleRequest
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

	// Register user
	user, err := h.handleConnect(initialMsg, ctx, conn)
	if err != nil {
		if verr, ok := err.(*utils.VerificationError); ok {
			sendAndLogError(ctx, conn, err, initialMsg, utils.InvalidFields, verr.UserError)
		}
		sendAndLogError(ctx, conn, err, initialMsg, utils.InvalidFields,
			map[string]string{"type": "Error creating user"})
		return
	}

	// Now handle messages
	h.processMessages(conn, ctx, user)
}

func (h *Handler) processMessages(conn *websocket.Conn, ctx context.Context, user User) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go h.pingConnection(ctx, conn)

	for {
		var msg Message

		// Use wsjson.Read to read JSON messages
		err := wsjson.Read(ctx, conn, &msg)
		if err != nil {
			// Check if it's a normal closure
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				log.Info().Str("username", user.Username).Msg("User closed connection normally")
			} else {
				log.Err(err).Str("username", user.Username).Msg("User disconnected")
			}
			return
		}

		switch msg.Type {
		case CLIENT_MESSAGE_TYPE.Connect:
			sendAndLogError(ctx, conn, fmt.Errorf("Already connected"), msg, utils.InvalidFields,
				map[string]string{"type": "Already connected"})
		case CLIENT_MESSAGE_TYPE.Status:

		case CLIENT_MESSAGE_TYPE.Attack:

		case CLIENT_MESSAGE_TYPE.ChangePokemon:

		case CLIENT_MESSAGE_TYPE.Surrender:

		default:
			sendAndLogError(ctx, conn, fmt.Errorf("Unknown message type"), msg, utils.InvalidFields,
				map[string]string{"received_type": fmt.Sprint(msg.Type)})
		}
	}
}

// Check connection every 30 seconds, if connection does not answer close.
func (h *Handler) pingConnection(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Send a ping, if it fails, connection is dead
			err := conn.Ping(ctx)
			if err != nil {
				log.Printf("Ping failed for user %v", err)
				conn.Close(websocket.StatusAbnormalClosure, "Ping failed")
				return
			}
		}
	}
}
