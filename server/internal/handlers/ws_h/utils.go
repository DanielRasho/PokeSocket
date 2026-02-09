package ws_h

import (
	"context"
	"encoding/json"

	"github.com/DanielRasho/PokeSocket/utils"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/rs/zerolog/log"
)

var CLIENT_MESSAGE_TYPE = struct {
	Connect       int
	Attack        int
	ChangePokemon int
	Surrender     int
	Status        int
	Match         int
}{
	Connect:       1,
	Attack:        2,
	ChangePokemon: 3,
	Surrender:     4,
	Status:        5,
	Match:         6,
}

var SERVER_MESSAGE_TYPE = struct {
	AcceptConnection int
	Attack           int
	ChangePokemon    int
	Status           int
	BattleEnded      int
	Disconnect       int
	Error            int
	MatchFound       int
	QueueJoined      int
}{
	AcceptConnection: 50,
	Attack:           51,
	ChangePokemon:    52,
	Status:           53,
	BattleEnded:      54,
	Disconnect:       55,
	Error:            56,
	MatchFound:       57,
	QueueJoined:      58,
}

// Helper function to create a message with any payload
func NewMessage(msgType int, payload interface{}) Message {
	if payload == nil {
		return Message{
			Type:    msgType,
			Payload: nil,
		}
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		panic("Could not create new websocket message")
	}

	return Message{
		Type:    msgType,
		Payload: json.RawMessage(payloadBytes),
	}
}

type ErrorResponse struct {
	Message string            `json:"msg"`
	Code    int               `json:"code"`
	Details map[string]string `json:"details"`
}

func (e *ErrorResponse) Error() string {
	return e.Message
}

func sendAndLogError(
	ctx context.Context,
	conn *websocket.Conn,
	err error, // for logging
	request Message,
	response *utils.DefaultMsg, // for sending
	details map[string]string) { // for sending

	msg := NewMessage(SERVER_MESSAGE_TYPE.Error, ErrorResponse{
		Message: response.Message,
		Code:    response.StatusCode,
		Details: details,
	})

	LogErrorRequest(err, request.Type, response.StatusCode, request.Payload, msg.Payload)

	wsjson.Write(ctx, conn, msg)
}

// Logs detailed information of an exception on the API when called.
// Parameters:
//   - err : Error that caused the exception
//   - body : Initial JSON payload that was sent by the client
func LogErrorRequest(err error, action int, statusCode int, body []byte, response []byte) {
	event := log.Error().Err(err).
		Int("status", statusCode).
		Int("action", action)
	if len(body) > 0 {
		event.RawJSON("body", body)
	}
	if len(response) > 0 {
		event.RawJSON("response", response)
	}
	event.Msg("Request failed")
}
