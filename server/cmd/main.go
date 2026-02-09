package main

import (
	"context"
	"net/http"

	"github.com/DanielRasho/PokeSocket/internal/config"
	"github.com/DanielRasho/PokeSocket/internal/handlers/http_h"
	"github.com/DanielRasho/PokeSocket/internal/handlers/ws_h"
	poke_mw "github.com/DanielRasho/PokeSocket/internal/middlewares"
	"github.com/DanielRasho/PokeSocket/internal/services/matchmaking_s"
	"github.com/DanielRasho/PokeSocket/internal/services/users_s"
	"github.com/DanielRasho/PokeSocket/internal/storage/postgres_cli"
	"github.com/DanielRasho/PokeSocket/internal/storage/postgres_cli/game_db"
	"github.com/DanielRasho/PokeSocket/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {

	// Load config
	godotenv.Load()
	apiPort := config.LoadRunningModeConfig()
	DBConfig := config.LoadDBConfig()
	loggingConfig := config.LoadLoggingConfig()
	corsConfig := config.LoadCorsConfig()

	// Config logger
	utils.ConfigureLogger(&loggingConfig)

	ctx := context.Background()

	// Start DB Connection
	DBCli, err := postgres_cli.NewPGClient(ctx, DBConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize UMS Postgres client")
		return
	}
	defer DBCli.Close()

	// ROUTER
	r := chi.NewRouter()

	// MIDDLEWARES
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(poke_mw.CreateCors(&corsConfig))

	// ROUTES
	api := newAPI(DBCli)
	r.Get("/health", api.checkHealth)
	r.Get("/pokemon", api.checkHealth)
	r.Get("/battle/stats", api.checkHealth)
	r.Get("/battle", api.battle)

	// Start server
	log.Printf("Running on http %s", apiPort)
	log.Fatal().Err(http.ListenAndServe(":"+apiPort, r))
}

type api struct {
	// HTTTP
	checkHealth http.HandlerFunc
	getPokemons http.HandlerFunc
	getStats    http.HandlerFunc

	// WS
	battle http.HandlerFunc
}

func newAPI(dbCli *pgxpool.Pool) api {
	validator := validator.New()
	DbQueries := game_db.New(dbCli)

	userService := users_s.UserService{
		DBClient:  dbCli,
		DBQueries: DbQueries,
	}

	// Create matchmaking service
	matchmakingService := matchmaking_s.NewMatchmakingService()

	return api{
		checkHealth: http_h.GetHealth,
		battle:      ws_h.NewHandler(dbCli, validator, &userService, matchmakingService),
	}
}
