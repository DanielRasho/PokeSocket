package users_s

import (
	"context"
	"fmt"

	"github.com/DanielRasho/PokeSocket/internal/storage/postgres_cli/game_db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type UserService struct {
	DBClient  *pgxpool.Pool
	DBQueries *game_db.Queries
}

func New(usersDBClient *pgxpool.Pool, usersQueries *game_db.Queries) *UserService {
	return &UserService{
		DBClient:  usersDBClient,
		DBQueries: usersQueries,
	}
}

func (s *UserService) CreateUserWithTeam(ctx context.Context, userId pgtype.UUID, username string, pokemonIds []int) error {
	tx, err := s.DBClient.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.DBQueries.WithTx(tx)

	// Insert user
	_, err = qtx.InsertUser(ctx, game_db.InsertUserParams{
		ID:       userId,
		Username: username,
	})
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	for position, pokemonId := range pokemonIds {
		err = qtx.InsertUserTeamPokemon(ctx, game_db.InsertUserTeamPokemonParams{
			UserID:           userId,
			PokemonSpeciesID: pgtype.Int4{Int32: int32(pokemonId), Valid: true},
			Position:         int32(position + 1),
		})
		if err != nil {
			log.Error().
				Err(err).
				Str("user_id", userId.String()).
				Int("pokemon_id", pokemonId).
				Int("position", position+1).
				Msg("Failed to insert pokemon to team")
			return fmt.Errorf("failed to insert pokemon %d: %w", pokemonId, err)
		}
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().
		Str("user_id", userId.String()).
		Str("username", username).
		Ints("pokemon_ids", pokemonIds).
		Msg("User and team created successfully")

	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, userId pgtype.UUID) error {
	return s.DBQueries.DeleteUser(ctx, userId)
}

type User struct {
	Id       pgtype.UUID
	Pokemons []int
}

func (s *UserService) GetUser(ctx context.Context, userId pgtype.UUID) error {
	return s.DBQueries.DeleteUser(ctx, userId)
}
