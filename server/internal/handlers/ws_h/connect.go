package ws_h

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DanielRasho/PokeSocket/utils"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type ClientConnectRequest struct {
	Username string `json:"username" validate:"required"`
}

type ClientConnectResponse struct {
	Username string    `json:"username"`
	Id       uuid.UUID `json:"uuid"`
}

func (h *Handler) handleConnect(msg Message, ctx context.Context, conn *websocket.Conn) (User, error) {

	var payload ClientConnectRequest
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return User{}, fmt.Errorf("")
	}

	if details, err := utils.ValidateStruct(&h.Validator, payload); err != nil {
		return User{}, &utils.VerificationError{
			Err:       err,
			UserError: details,
			Code:      utils.InvalidFields,
		}
	}

	id := uuid.New()

	log.Debug().Str("Username", payload.Username).Str("Id", id.String())

	err := wsjson.Write(ctx, conn, NewMessage(SERVER_MESSAGE_TYPE.AcceptConnection, ClientConnectResponse{
		Username: payload.Username,
		Id:       id,
	}))
	if err != nil {
		return User{}, nil
	}

	return User{
		Username: payload.Username,
		Id:       id,
	}, nil
}
