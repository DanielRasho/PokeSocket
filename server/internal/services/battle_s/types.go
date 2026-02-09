package battle

import (
	"github.com/DanielRasho/PokeSocket/internal/storage/postgres_cli/game_db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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

// TODO: player1, player2, both players pokemons starts with full health
func (s *BattleService) CreateBattle() (pgtype.UUID, error) {

	return pgtype.UUID{}, nil
}
