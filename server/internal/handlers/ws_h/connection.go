package ws_h

import (
	"context"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/rs/zerolog/log"
)

const SESSION_WRITE_BUFFER_SIZE = 4

type Connection struct {
	PlayerID PlayerID
	Username string
	Pokemons []int
	Conn     *websocket.Conn
	Send     chan Message // Buffered channel for async sends
	Ctx      context.Context
	Cancel   context.CancelFunc
}

// Close gracefully closes the connection
func (c *Connection) Close() {
	c.Cancel()    // Cancel context to stop goroutines
	close(c.Send) // Close channel to signal writePump to stop
}

func NewConnection(playerID PlayerID, username string, pokemons []int, conn *websocket.Conn, parentCtx context.Context) *Connection {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Connection{
		PlayerID: playerID,
		Username: username,
		Pokemons: pokemons,
		Conn:     conn,
		Send:     make(chan Message, SESSION_WRITE_BUFFER_SIZE),
		Ctx:      ctx,
		Cancel:   cancel,
	}
}

// pingConnection periodically pings to keep connection alive
func (conn *Connection) pingConnection() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-conn.Ctx.Done():
			log.Debug().Str("username", conn.Username).Msg("Ping routine stopped")
			return
		case <-ticker.C:
			err := conn.Conn.Ping(conn.Ctx)
			if err != nil {
				log.Warn().
					Err(err).
					Str("username", conn.Username).
					Msg("Ping failed, closing connection")
				conn.Conn.Close(websocket.StatusAbnormalClosure, "Ping failed")
				conn.Close() // Trigger cleanup
				return
			}
		}
	}
}

// writePump sends queued messages from the buffered channel to WebSocket
func (conn *Connection) writePump() {
	defer func() {
		log.Info().Str("username", conn.Username).Msg("Write pump stopped")
	}()

	for {
		select {
		case <-conn.Ctx.Done():
			return
		case message, ok := <-conn.Send:
			if !ok {
				conn.Conn.Close(websocket.StatusNormalClosure, "Channel closed")
				return
			}

			err := wsjson.Write(conn.Ctx, conn.Conn, message)
			if err != nil {
				log.Err(err).Str("username", conn.Username).Msg("Error writing message")
				return
			}
		}
	}
}
