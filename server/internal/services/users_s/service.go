package users_s

import (
	"context"

	"github.com/DanielRasho/PokeSocket/internal/storage/postgres_cli/game_db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (s *UserService) InsertUser(ctx context.Context, username string) (pgtype.UUID, error) {
	return s.DBQueries.InsertUser(ctx, username)
}

func (s *UserService) DeleteUser(ctx context.Context, userId pgtype.UUID) error {
	return s.DBQueries.DeleteUser(ctx, userId)
}
