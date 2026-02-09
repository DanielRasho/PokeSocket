package ws_h

import (
	"context"
	"encoding/json"

	"github.com/DanielRasho/PokeSocket/internal/services/battle_s"
	"github.com/DanielRasho/PokeSocket/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

type AttackRequestPayload struct {
	BattleID string `json:"battle_id" validate:"required,uuid"`
	MoveID   int    `json:"move_id" validate:"required"`
}

type BattleStateResponse struct {
	BattleID     string           `json:"battle_id"`
	Message      string           `json:"message"`
	YourInfo     PlayerBattleInfo `json:"your_info"`
	OpponentInfo PlayerBattleInfo `json:"opponent_info"`
	BattleEnded  bool             `json:"battle_ended,omitempty"`
	Winner       string           `json:"winner,omitempty"` // player_id of winner
}

// handleAttack processes an attack action in a battle
func (h *Handler) handleAttack(conn *Connection, msg Message) {
	var payload AttackRequestPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		sendAndLogError(conn.Ctx, conn.Conn, err, msg, utils.InvalidFields,
			map[string]string{"error": "Invalid payload"})
		return
	}

	ctx := context.Background()

	// Parse battle ID
	var battleUUID pgtype.UUID
	if scanErr := battleUUID.Scan(payload.BattleID); scanErr != nil {
		sendAndLogError(conn.Ctx, conn.Conn, scanErr, msg, utils.BadRequest,
			map[string]string{"error": "Invalid UUID format"})
		return
	}

	// Get opponent ID from connections map
	// TODO: Better way to get opponent ID from battle
	var opponentID pgtype.UUID
	for playerID := range h.Connections {
		if playerID != conn.PlayerID {
			opponentID = playerID
			break
		}
	}

	// Process the attack
	attackReq := battle_s.AttackRequest{
		BattleID:   battleUUID,
		AttackerID: conn.PlayerID,
		DefenderID: opponentID,
		MoveID:     payload.MoveID,
	}

	battleState, err := h.BattleService.AttackPokemon(ctx, attackReq)
	if err != nil {
		sendAndLogError(conn.Ctx, conn.Conn, err, msg, utils.BadRequest,
			map[string]string{"error": err.Error()})
		return
	}

	// Convert teams to PokemonInfo
	player1Team := make([]PokemonInfo, len(battleState.Player1Team))
	for i, poke := range battleState.Player1Team {
		player1Team[i] = PokemonInfo{
			SpeciesID: int(poke.PokemonSpeciesID.Int32),
			Position:  poke.Position,
			CurrentHP: poke.CurrentHp,
			IsFainted: poke.IsFainted.Bool,
		}
	}

	player2Team := make([]PokemonInfo, len(battleState.Player2Team))
	for i, poke := range battleState.Player2Team {
		player2Team[i] = PokemonInfo{
			SpeciesID: int(poke.PokemonSpeciesID.Int32),
			Position:  poke.Position,
			CurrentHP: poke.CurrentHp,
			IsFainted: poke.IsFainted.Bool,
		}
	}

	// Determine which player is player1 and which is player2
	var yourTeam, opponentTeam []PokemonInfo
	var yourID, opponentPlayerID pgtype.UUID
	var yourActivePos, opponentActivePos int32

	if conn.PlayerID == battleState.Player1ID {
		yourTeam = player1Team
		opponentTeam = player2Team
		yourID = battleState.Player1ID
		opponentPlayerID = battleState.Player2ID
		yourActivePos = battleState.Player1ActivePos
		opponentActivePos = battleState.Player2ActivePos
	} else {
		yourTeam = player2Team
		opponentTeam = player1Team
		yourID = battleState.Player2ID
		opponentPlayerID = battleState.Player1ID
		yourActivePos = battleState.Player2ActivePos
		opponentActivePos = battleState.Player1ActivePos
	}

	// Get opponent connection for username
	h.mu.RLock()
	opponentConn, opponentExists := h.Connections[opponentPlayerID]
	h.mu.RUnlock()

	if !opponentExists {
		log.Error().Str("opponent_id", opponentPlayerID.String()).Msg("Opponent not found")
		return
	}

	// Create response for attacker
	attackerResponse := BattleStateResponse{
		BattleID: battleState.BattleID.String(),
		Message:  battleState.Message,
		YourInfo: PlayerBattleInfo{
			PlayerID:      yourID.String(),
			Username:      conn.Username,
			Team:          yourTeam,
			ActivePokemon: yourActivePos,
		},
		OpponentInfo: PlayerBattleInfo{
			PlayerID:      opponentPlayerID.String(),
			Username:      opponentConn.Username,
			Team:          opponentTeam,
			ActivePokemon: opponentActivePos,
		},
		BattleEnded: battleState.BattleEnded,
	}

	if battleState.BattleEnded {
		attackerResponse.Winner = battleState.WinnerID.String()
	}

	// Create response for defender (swap your/opponent)
	defenderResponse := BattleStateResponse{
		BattleID: battleState.BattleID.String(),
		Message:  battleState.Message,
		YourInfo: PlayerBattleInfo{
			PlayerID:      opponentPlayerID.String(),
			Username:      opponentConn.Username,
			Team:          opponentTeam,
			ActivePokemon: opponentActivePos,
		},
		OpponentInfo: PlayerBattleInfo{
			PlayerID:      yourID.String(),
			Username:      conn.Username,
			Team:          yourTeam,
			ActivePokemon: yourActivePos,
		},
		BattleEnded: battleState.BattleEnded,
	}

	if battleState.BattleEnded {
		defenderResponse.Winner = battleState.WinnerID.String()
	}

	// Send to both players
	conn.Send <- NewMessage(SERVER_MESSAGE_TYPE.Attack, attackerResponse)

	err = h.SendToPlayer(opponentPlayerID, NewMessage(SERVER_MESSAGE_TYPE.Attack, defenderResponse))
	if err != nil {
		log.Warn().Err(err).Str("opponent_id", opponentPlayerID.String()).Msg("Failed to notify opponent")
	}

	// If battle ended, delete it
	if battleState.BattleEnded {

		// Send to both players the battle ended
		conn.Send <- NewMessage(SERVER_MESSAGE_TYPE.BattleEnded, attackerResponse)
		err = h.SendToPlayer(opponentPlayerID, NewMessage(SERVER_MESSAGE_TYPE.Attack, defenderResponse))
		if err != nil {
			log.Warn().Err(err).Str("opponent_id", opponentPlayerID.String()).Msg("Failed to notify opponent")
		}

		if deleteErr := h.BattleService.DeleteBattle(ctx, battleUUID); deleteErr != nil {
			log.Error().
				Err(deleteErr).
				Str("battle_id", payload.BattleID).
				Msg("Failed to delete battle after completion")
		} else {
			log.Info().
				Str("battle_id", payload.BattleID).
				Str("winner_id", battleState.WinnerID.String()).
				Msg("Battle completed and deleted")
		}
	}

	log.Info().
		Str("player_id", conn.PlayerID.String()).
		Str("battle_id", payload.BattleID).
		Int("move_id", payload.MoveID).
		Bool("battle_ended", battleState.BattleEnded).
		Msg("Attack processed and state sent to both players")
}
