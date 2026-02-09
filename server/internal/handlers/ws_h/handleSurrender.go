package ws_h

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

type SurrenderRequest struct {
	BattleID string `json:"battle_id" validate:"required,uuid"`
}

// TODO: To implement
func (h *Handler) handleSurrender(conn *Connection, msg Message) {
	var payload SurrenderRequest
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Error().Err(err).Msg("Failed to parse surrender request")
		conn.Send <- NewMessage(SERVER_MESSAGE_TYPE.Error, ErrorResponse{
			Message: "Invalid surrender request",
			Code:    400,
			Details: map[string]string{"error": "Invalid payload"},
		})
		return
	}

	// TODO: Implement surrender logic
	// - Validate battle exists
	// - Verify player is in the battle
	// - Set winner as the opponent
	// - Notify both players of the result

	ctx := context.Background()

	// Parse battle ID
	var battleUUID pgtype.UUID
	if scanErr := battleUUID.Scan(payload.BattleID); scanErr != nil {
		log.Error().Err(scanErr).Msg("Invalid battle UUID")
		conn.Send <- NewMessage(SERVER_MESSAGE_TYPE.Error, ErrorResponse{
			Message: "Invalid battle ID",
			Code:    400,
			Details: map[string]string{"error": "Invalid UUID format"},
		})
		return
	}

	// DELETE BATTLE FROM DATABASE
	if deleteErr := h.BattleService.DeleteBattle(ctx, battleUUID); deleteErr != nil {
		log.Error().
			Err(deleteErr).
			Str("battle_id", payload.BattleID).
			Msg("Failed to delete battle after surrender")
	}

	log.Info().
		Str("player_id", conn.PlayerID.String()).
		Str("battle_id", payload.BattleID).
		Msg("Player surrendered, battle deleted")
}
