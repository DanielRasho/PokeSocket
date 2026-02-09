package ws_h

import (
	"context"
	"encoding/json"

	"github.com/DanielRasho/PokeSocket/internal/services/battle_s"
	"github.com/DanielRasho/PokeSocket/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

// ChangePokemonRequestPayload contains the pokemon switch details from client
type ChangePokemonRequestPayload struct {
	BattleID string `json:"battle_id" validate:"required,uuid"`
	Position int32  `json:"position" validate:"required"`
}

// handleChangePokemon processes a pokemon switch action in a battle
func (h *Handler) handleChangePokemon(conn *Connection, msg Message) {
	var payload ChangePokemonRequestPayload
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

	// Process the switch
	switchReq := battle_s.SwitchPokemonRequest{
		BattleID:    battleUUID,
		PlayerID:    conn.PlayerID,
		OpponentID:  opponentID,
		NewPosition: payload.Position,
	}

	battleState, err := h.BattleService.SwitchPokemon(ctx, switchReq)
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

	if conn.PlayerID == battleState.Player1ID {
		yourTeam = player1Team
		opponentTeam = player2Team
		yourID = battleState.Player1ID
		opponentPlayerID = battleState.Player2ID
	} else {
		yourTeam = player2Team
		opponentTeam = player1Team
		yourID = battleState.Player2ID
		opponentPlayerID = battleState.Player1ID
	}

	// Get opponent connection for username
	h.mu.RLock()
	opponentConn, opponentExists := h.Connections[opponentPlayerID]
	h.mu.RUnlock()

	if !opponentExists {
		log.Error().Str("opponent_id", opponentPlayerID.String()).Msg("Opponent not found")
		return
	}

	// Create response for switcher
	switcherResponse := BattleStateResponse{
		BattleID: battleState.BattleID.String(),
		Message:  battleState.Message,
		YourInfo: PlayerBattleInfo{
			PlayerID: yourID.String(),
			Username: conn.Username,
			Team:     yourTeam,
		},
		OpponentInfo: PlayerBattleInfo{
			PlayerID: opponentPlayerID.String(),
			Username: opponentConn.Username,
			Team:     opponentTeam,
		},
		BattleEnded: battleState.BattleEnded,
	}

	if battleState.BattleEnded {
		switcherResponse.Winner = battleState.WinnerID.String()
	}

	// Create response for opponent (swap your/opponent)
	opponentResponse := BattleStateResponse{
		BattleID: battleState.BattleID.String(),
		Message:  battleState.Message,
		YourInfo: PlayerBattleInfo{
			PlayerID: opponentPlayerID.String(),
			Username: opponentConn.Username,
			Team:     opponentTeam,
		},
		OpponentInfo: PlayerBattleInfo{
			PlayerID: yourID.String(),
			Username: conn.Username,
			Team:     yourTeam,
		},
		BattleEnded: battleState.BattleEnded,
	}

	if battleState.BattleEnded {
		opponentResponse.Winner = battleState.WinnerID.String()
	}

	// Send to both players (use ChangePokemon message type)
	conn.Send <- NewMessage(SERVER_MESSAGE_TYPE.ChangePokemon, switcherResponse)

	err = h.SendToPlayer(opponentPlayerID, NewMessage(SERVER_MESSAGE_TYPE.ChangePokemon, opponentResponse))
	if err != nil {
		log.Warn().Err(err).Str("opponent_id", opponentPlayerID.String()).Msg("Failed to notify opponent")
	}

	log.Info().
		Str("player_id", conn.PlayerID.String()).
		Str("battle_id", payload.BattleID).
		Int32("new_position", payload.Position).
		Msg("Pokemon switch processed and state sent to both players")
}
