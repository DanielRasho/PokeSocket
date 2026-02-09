package ws_h

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DanielRasho/PokeSocket/utils"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

type ClientConnectRequest struct {
	Username string `json:"username" validate:"required"`
	Pokemons []int  `json:"pokemons" validate:"required,len=3"`
}

type ClientConnectResponse struct {
	Username string      `json:"username"`
	Id       pgtype.UUID `json:"id"`
}

func (h *Handler) handleConnect(msg Message, ctx context.Context, conn *websocket.Conn) (*Connection, error) {

	var payload ClientConnectRequest
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, fmt.Errorf("")
	}

	if details, err := utils.ValidateStruct(&h.Validator, payload); err != nil {
		return nil, &utils.VerificationError{
			Err:       err,
			UserError: details,
			Code:      utils.InvalidFields,
		}
	}

	userId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	log.Debug().Str("Username", payload.Username).Str("Id", userId.String()).Msg("Creating user")

	// CREATE USER AND TEAM IN DATABASE
	err := h.UserService.CreateUserWithTeam(ctx, userId, payload.Username, payload.Pokemons)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user and team")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// SEND SESSION DATA TO CLIENT.
	err = wsjson.Write(ctx, conn, NewMessage(
		SERVER_MESSAGE_TYPE.AcceptConnection,
		ClientConnectResponse{
			Username: payload.Username,
			Id:       userId,
		}))
	if err != nil {
		return nil, nil
	}

	return NewConnection(userId, payload.Username, payload.Pokemons, conn, ctx), nil
}
