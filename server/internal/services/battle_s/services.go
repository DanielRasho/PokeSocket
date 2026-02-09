package battle_s

import (
	"context"
	"fmt"

	"github.com/DanielRasho/PokeSocket/internal/storage/postgres_cli/game_db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type BattleService struct {
	DBClient  *pgxpool.Pool
	DBQueries *game_db.Queries
}

func New(usersDBClient *pgxpool.Pool, usersQueries *game_db.Queries) *BattleService {
	return &BattleService{
		DBClient:  usersDBClient,
		DBQueries: usersQueries,
	}
}

// BattleInfo contains all information about a created battle
type BattleInfo struct {
	BattleID    pgtype.UUID
	Player1ID   pgtype.UUID
	Player1Team []game_db.UserTeam
	Player2ID   pgtype.UUID
	Player2Team []game_db.UserTeam
}

// CreateBattle creates a new battle between two players
func (s *BattleService) CreateBattle(ctx context.Context, player1ID, player2ID pgtype.UUID) (*BattleInfo, error) {
	// Generate battle ID
	battleID := pgtype.UUID{Bytes: uuid.New(), Valid: true}

	// Create battle entry
	battle, err := s.DBQueries.CreateBattle(ctx, game_db.CreateBattleParams{
		ID:        battleID,
		Player1ID: player1ID,
		Player2ID: player2ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create battle: %w", err)
	}

	// Get player 1 team
	player1Team, err := s.DBQueries.GetUserTeam(ctx, player1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player1 team: %w", err)
	}

	// Get player 2 team
	player2Team, err := s.DBQueries.GetUserTeam(ctx, player2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player2 team: %w", err)
	}

	log.Info().
		Str("battle_id", battle.ID.String()).
		Str("player1_id", player1ID.String()).
		Str("player2_id", player2ID.String()).
		Msg("Battle created successfully")

	return &BattleInfo{
		BattleID:    battle.ID,
		Player1ID:   player1ID,
		Player1Team: player1Team,
		Player2ID:   player2ID,
		Player2Team: player2Team,
	}, nil
}

// DeleteBattle removes a battle from the database
func (s *BattleService) DeleteBattle(ctx context.Context, battleID pgtype.UUID) error {
	err := s.DBQueries.DeleteBattle(ctx, battleID)
	if err != nil {
		return fmt.Errorf("failed to delete battle: %w", err)
	}

	log.Info().
		Str("battle_id", battleID.String()).
		Msg("Battle deleted successfully")

	return nil
}

// AttackRequest contains all data needed for an attack
type AttackRequest struct {
	BattleID   pgtype.UUID
	AttackerID pgtype.UUID
	DefenderID pgtype.UUID
	MoveID     int
}

// BattleStateResult contains the complete battle state after an action
type BattleStateResult struct {
	BattleID    pgtype.UUID
	Message     string
	Player1ID   pgtype.UUID
	Player1Team []game_db.UserTeam
	Player2ID   pgtype.UUID
	Player2Team []game_db.UserTeam
	BattleEnded bool
	WinnerID    pgtype.UUID
}

// AttackPokemon processes a pokemon attack and returns the new battle state
func (s *BattleService) AttackPokemon(ctx context.Context, req AttackRequest) (*BattleStateResult, error) {
	// Get battle info
	battle, err := s.DBQueries.GetBattle(ctx, req.BattleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get battle: %w", err)
	}

	// Validate turn: odd turns = player1, even turns = player2
	currentTurn := battle.CurrentTurn.Int32
	isPlayer1Turn := currentTurn%2 == 1
	isPlayer1Attacking := req.AttackerID.Bytes == battle.Player1ID.Bytes

	if isPlayer1Turn && !isPlayer1Attacking {
		return nil, fmt.Errorf("not your turn - it's player1's turn")
	}
	if !isPlayer1Turn && isPlayer1Attacking {
		return nil, fmt.Errorf("not your turn - it's player2's turn")
	}

	// Determine defender's active pokemon position based on who is attacking
	var defenderPos int32
	if req.AttackerID.Bytes == battle.Player1ID.Bytes {
		defenderPos = battle.Player2ActivePokemonPosition.Int32
	} else {
		defenderPos = battle.Player1ActivePokemonPosition.Int32
	}

	// Get move information (for now, using a simple damage calculation)
	// TODO: Get actual move data from pokemon_moves table
	baseDamage := 10 // Placeholder damage

	// Calculate damage (simplified formula)
	damage := int32(baseDamage)

	// Get defender's current pokemon HP
	defenderTeam, err := s.DBQueries.GetUserTeam(ctx, req.DefenderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get defender team: %w", err)
	}

	// Find the defending pokemon
	var defenderPokemon game_db.UserTeam
	for _, poke := range defenderTeam {
		if poke.Position == defenderPos {
			defenderPokemon = poke
			break
		}
	}

	// Calculate new HP
	newHP := defenderPokemon.CurrentHp - damage
	if newHP < 0 {
		newHP = 0
	}

	// Update defender's pokemon HP
	err = s.DBQueries.UpdatePokemonHP(ctx, game_db.UpdatePokemonHPParams{
		UserID:    req.DefenderID,
		Position:  defenderPos,
		CurrentHp: newHP,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update pokemon HP: %w", err)
	}

	// If pokemon fainted, auto-switch to next available pokemon
	if newHP == 0 {
		nextPokemon := s.findNextAvailablePokemon(defenderTeam, defenderPos)
		if nextPokemon != nil {
			// Update the battle's active pokemon position for the defender
			if req.DefenderID.Bytes == battle.Player1ID.Bytes {
				err = s.DBQueries.UpdatePlayer1ActivePokemon(ctx, game_db.UpdatePlayer1ActivePokemonParams{
					ID:                           battle.ID,
					Player1ActivePokemonPosition: pgtype.Int4{Int32: nextPokemon.Position, Valid: true},
				})
			} else {
				err = s.DBQueries.UpdatePlayer2ActivePokemon(ctx, game_db.UpdatePlayer2ActivePokemonParams{
					ID:                           battle.ID,
					Player2ActivePokemonPosition: pgtype.Int4{Int32: nextPokemon.Position, Valid: true},
				})
			}
			if err != nil {
				log.Warn().Err(err).Msg("Failed to update active pokemon position")
			}
		}
	}

	// Update battle turn
	err = s.DBQueries.UpdateBattleTurn(ctx, req.BattleID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to update battle turn")
	}

	// Get updated teams
	player1Team, err := s.DBQueries.GetUserTeam(ctx, battle.Player1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player1 team: %w", err)
	}

	player2Team, err := s.DBQueries.GetUserTeam(ctx, battle.Player2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player2 team: %w", err)
	}

	// Create message describing the action
	isFainted := newHP == 0
	message := fmt.Sprintf("Attack dealt %d damage! Defender's HP: %d", damage, newHP)
	if isFainted {
		message += " - Pokemon fainted!"
	}

	// Check if battle is over (all pokemon fainted)
	battleEnded, winnerID := s.checkBattleEnd(player1Team, player2Team, battle.Player1ID, battle.Player2ID)

	log.Info().
		Str("battle_id", req.BattleID.String()).
		Str("attacker_id", req.AttackerID.String()).
		Int32("damage", damage).
		Int32("new_hp", newHP).
		Bool("fainted", isFainted).
		Bool("battle_ended", battleEnded).
		Msg("Attack processed")

	return &BattleStateResult{
		BattleID:    req.BattleID,
		Message:     message,
		Player1ID:   battle.Player1ID,
		Player1Team: player1Team,
		Player2ID:   battle.Player2ID,
		Player2Team: player2Team,
		BattleEnded: battleEnded,
		WinnerID:    winnerID,
	}, nil
}

// checkBattleEnd checks if all pokemon of one team are fainted
func (s *BattleService) checkBattleEnd(player1Team, player2Team []game_db.UserTeam, player1ID, player2ID pgtype.UUID) (bool, pgtype.UUID) {
	player1AllFainted := true
	player2AllFainted := true

	for _, poke := range player1Team {
		if !poke.IsFainted.Bool {
			player1AllFainted = false
			break
		}
	}

	for _, poke := range player2Team {
		if !poke.IsFainted.Bool {
			player2AllFainted = false
			break
		}
	}

	if player1AllFainted {
		return true, player2ID
	}
	if player2AllFainted {
		return true, player1ID
	}

	return false, pgtype.UUID{}
}

// findNextAvailablePokemon finds the next pokemon with HP > 0 after the current position
func (s *BattleService) findNextAvailablePokemon(team []game_db.UserTeam, currentPos int32) *game_db.UserTeam {
	for _, poke := range team {
		if poke.Position != currentPos && poke.CurrentHp > 0 && !poke.IsFainted.Bool {
			return &poke
		}
	}
	return nil
}

// SwitchPokemonRequest contains all data needed for switching pokemon
type SwitchPokemonRequest struct {
	BattleID   pgtype.UUID
	PlayerID   pgtype.UUID
	OpponentID pgtype.UUID
	NewPosition int32
}

// SwitchPokemon handles switching the active pokemon for a player
func (s *BattleService) SwitchPokemon(ctx context.Context, req SwitchPokemonRequest) (*BattleStateResult, error) {
	// Get battle info
	battle, err := s.DBQueries.GetBattle(ctx, req.BattleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get battle: %w", err)
	}

	// Validate turn: odd turns = player1, even turns = player2
	currentTurn := battle.CurrentTurn.Int32
	isPlayer1Turn := currentTurn%2 == 1
	isPlayer1Switching := req.PlayerID.Bytes == battle.Player1ID.Bytes

	if isPlayer1Turn && !isPlayer1Switching {
		return nil, fmt.Errorf("not your turn - it's player1's turn")
	}
	if !isPlayer1Turn && isPlayer1Switching {
		return nil, fmt.Errorf("not your turn - it's player2's turn")
	}

	// Get player's team
	playerTeam, err := s.DBQueries.GetUserTeam(ctx, req.PlayerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player team: %w", err)
	}

	// Find the pokemon at the requested position
	var targetPokemon *game_db.UserTeam
	for _, poke := range playerTeam {
		if poke.Position == req.NewPosition {
			targetPokemon = &poke
			break
		}
	}

	if targetPokemon == nil {
		return nil, fmt.Errorf("no pokemon at position %d", req.NewPosition)
	}

	// Check if pokemon is fainted
	if targetPokemon.IsFainted.Bool || targetPokemon.CurrentHp <= 0 {
		return nil, fmt.Errorf("cannot switch to a fainted pokemon")
	}

	// Check if already active
	var currentActivePos int32
	if isPlayer1Switching {
		currentActivePos = battle.Player1ActivePokemonPosition.Int32
	} else {
		currentActivePos = battle.Player2ActivePokemonPosition.Int32
	}

	if currentActivePos == req.NewPosition {
		return nil, fmt.Errorf("pokemon is already active")
	}

	// Update the battle's active pokemon position
	if isPlayer1Switching {
		err = s.DBQueries.UpdatePlayer1ActivePokemon(ctx, game_db.UpdatePlayer1ActivePokemonParams{
			ID:                           battle.ID,
			Player1ActivePokemonPosition: pgtype.Int4{Int32: req.NewPosition, Valid: true},
		})
	} else {
		err = s.DBQueries.UpdatePlayer2ActivePokemon(ctx, game_db.UpdatePlayer2ActivePokemonParams{
			ID:                           battle.ID,
			Player2ActivePokemonPosition: pgtype.Int4{Int32: req.NewPosition, Valid: true},
		})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update active pokemon: %w", err)
	}

	// Advance turn
	err = s.DBQueries.UpdateBattleTurn(ctx, req.BattleID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to update battle turn")
	}

	// Get updated teams
	player1Team, err := s.DBQueries.GetUserTeam(ctx, battle.Player1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player1 team: %w", err)
	}

	player2Team, err := s.DBQueries.GetUserTeam(ctx, battle.Player2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player2 team: %w", err)
	}

	// Create message describing the action
	message := fmt.Sprintf("Switched to %s (position %d)", targetPokemon.PokemonSpeciesID.Int32, req.NewPosition)

	// Check if battle is over (shouldn't happen on switch, but check anyway)
	battleEnded, winnerID := s.checkBattleEnd(player1Team, player2Team, battle.Player1ID, battle.Player2ID)

	log.Info().
		Str("battle_id", req.BattleID.String()).
		Str("player_id", req.PlayerID.String()).
		Int32("new_position", req.NewPosition).
		Msg("Pokemon switched successfully")

	return &BattleStateResult{
		BattleID:    req.BattleID,
		Message:     message,
		Player1ID:   battle.Player1ID,
		Player1Team: player1Team,
		Player2ID:   battle.Player2ID,
		Player2Team: player2Team,
		BattleEnded: battleEnded,
		WinnerID:    winnerID,
	}, nil
}
