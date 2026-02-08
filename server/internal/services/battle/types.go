package battle

import (
	"time"

	"github.com/google/uuid"
)

type Pokemon struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	HP      int    `json:"hp"`
	MaxHP   int    `json:"max_hp"`
	Attack  int    `json:"attack"`
	Defense int    `json:"defense"`
	Speed   int    `json:"speed"`
	Moves   []Move `json:"moves"`
}

type Move struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Power    int    `json:"power"`
	Accuracy int    `json:"accuracy"`
	Type     string `json:"type"`
}

type Player struct {
	ID            uuid.UUID `json:"id"`
	Username      string    `json:"username"`
	Team          []Pokemon `json:"team"`
	ActivePokemon int       `json:"active_pokemon"` // index in team
}

type Battle struct {
	ID          uuid.UUID     `json:"id"`
	Player1     *Player       `json:"player1"`
	Player2     *Player       `json:"player2"`
	CurrentTurn uuid.UUID     `json:"current_turn"` // player ID
	Status      string        `json:"status"`       // "waiting", "active", "finished"
	Winner      *uuid.UUID    `json:"winner,omitempty"`
	BattleLog   []BattleEvent `json:"battle_log"`
	CreatedAt   time.Time     `json:"created_at"`
}

type BattleEvent struct {
	ID        int       `json:"id"`
	BattleID  uuid.UUID `json:"battle_id"`
	PlayerID  uuid.UUID `json:"player_id"`
	EventType string    `json:"event_type"` // "attack", "switch", "surrender", "damage", "faint"
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type Action struct {
	Type         string `json:"type"` // "attack", "switch", "surrender"
	MoveIndex    *int   `json:"move_index,omitempty"`
	PokemonIndex *int   `json:"pokemon_index,omitempty"`
}
