package ws_h

import "github.com/rs/zerolog/log"

type MatchFoundResponse struct {
	OpponentID       string `json:"opponent_id"`
	OpponentUsername string `json:"opponent_username"`
	OpponentPokemon  []int  `json:"oponnent_pokemon"`
}

type QueueJoinedResponse struct {
	Message   string `json:"message"`
	QueueSize int    `json:"queue_size"`
}

func (h *Handler) handleMatch(conn *Connection) {
	// Try to match the player
	opponent := h.MatchmakingService.EnterQueue(conn.PlayerID, conn.Username)

	// MATCH FOUND
	if opponent != nil {

		// NOTIFY BOTH PLAYERS
		conn.Send <- NewMessage(SERVER_MESSAGE_TYPE.MatchFound, MatchFoundResponse{
			OpponentID:       opponent.PlayerID.String(),
			OpponentUsername: opponent.Username,
		})

		err := h.SendToPlayer(opponent.PlayerID, NewMessage(SERVER_MESSAGE_TYPE.MatchFound, MatchFoundResponse{
			OpponentID:       conn.PlayerID.String(),
			OpponentUsername: conn.Username,
		}))

		if err != nil {
			log.Warn().
				Err(err).
				Str("opponent_id", opponent.PlayerID.String()).
				Msg("Failed to notify opponent of match")
		}
	} else {
		// No match found, player added to queue
		conn.Send <- NewMessage(SERVER_MESSAGE_TYPE.QueueJoined, QueueJoinedResponse{
			Message:   "Joined matchmaking queue, waiting for opponent...",
			QueueSize: h.MatchmakingService.GetQueueSize(),
		})
	}
}
