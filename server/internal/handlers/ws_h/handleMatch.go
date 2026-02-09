package ws_h

import (
	"context"

	"github.com/rs/zerolog/log"
)

type PokemonInfo struct {
	SpeciesID int   `json:"species_id"`
	Position  int32 `json:"position"`
	CurrentHP int32 `json:"current_hp"`
	IsFainted bool  `json:"is_fainted"`
}

type PlayerBattleInfo struct {
	PlayerID string        `json:"player_id"`
	Username string        `json:"username"`
	Team     []PokemonInfo `json:"team"`
}

type MatchFoundResponse struct {
	BattleID     string           `json:"battle_id"`
	YourInfo     PlayerBattleInfo `json:"your_info"`
	OpponentInfo PlayerBattleInfo `json:"opponent_info"`
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
		ctx := context.Background()

		// CREATE BATTLE IN DATABASE
		// Opponent is player1 (first in queue), conn is player2 (second in queue)
		battleInfo, err := h.BattleService.CreateBattle(ctx, opponent.PlayerID, conn.PlayerID)
		if err != nil {
			log.Error().
				Err(err).
				Str("player1_id", opponent.PlayerID.String()).
				Str("player2_id", conn.PlayerID.String()).
				Msg("Failed to create battle")

			// Send error to both players
			conn.Send <- NewMessage(SERVER_MESSAGE_TYPE.Error, ErrorResponse{
				Message: "Failed to create battle",
				Code:    500,
				Details: map[string]string{"error": "Could not start battle"},
			})
			return
		}

		// PREPARE MESSAGES FOR SENDING
		player1Team := make([]PokemonInfo, len(battleInfo.Player1Team))
		for i, poke := range battleInfo.Player1Team {
			player1Team[i] = PokemonInfo{
				SpeciesID: int(poke.PokemonSpeciesID.Int32),
				Position:  poke.Position,
				CurrentHP: poke.CurrentHp,
				IsFainted: poke.IsFainted.Bool,
			}
		}

		player2Team := make([]PokemonInfo, len(battleInfo.Player2Team))
		for i, poke := range battleInfo.Player2Team {
			player2Team[i] = PokemonInfo{
				SpeciesID: int(poke.PokemonSpeciesID.Int32),
				Position:  poke.Position,
				CurrentHP: poke.CurrentHp,
				IsFainted: poke.IsFainted.Bool,
			}
		}

		// SEND MESSAGES
		// opponent is player1, conn is player2

		conn.Send <- NewMessage(SERVER_MESSAGE_TYPE.MatchFound, MatchFoundResponse{
			BattleID: battleInfo.BattleID.String(),
			YourInfo: PlayerBattleInfo{
				PlayerID: conn.PlayerID.String(),
				Username: conn.Username,
				Team:     player2Team,
			},
			OpponentInfo: PlayerBattleInfo{
				PlayerID: opponent.PlayerID.String(),
				Username: opponent.Username,
				Team:     player1Team,
			},
		})

		err = h.SendToPlayer(opponent.PlayerID, NewMessage(SERVER_MESSAGE_TYPE.MatchFound, MatchFoundResponse{
			BattleID: battleInfo.BattleID.String(),
			YourInfo: PlayerBattleInfo{
				PlayerID: opponent.PlayerID.String(),
				Username: opponent.Username,
				Team:     player1Team,
			},
			OpponentInfo: PlayerBattleInfo{
				PlayerID: conn.PlayerID.String(),
				Username: conn.Username,
				Team:     player2Team,
			},
		}))

		if err != nil {
			log.Warn().
				Err(err).
				Str("opponent_id", opponent.PlayerID.String()).
				Msg("Failed to notify opponent of match")
		}

		log.Info().
			Str("battle_id", battleInfo.BattleID.String()).
			Str("player1", conn.Username).
			Str("player2", opponent.Username).
			Msg("Battle started and players notified")
	} else {
		// No match found, player added to queue
		conn.Send <- NewMessage(SERVER_MESSAGE_TYPE.QueueJoined, QueueJoinedResponse{
			Message:   "Joined matchmaking queue, waiting for opponent...",
			QueueSize: h.MatchmakingService.GetQueueSize(),
		})
	}
}
