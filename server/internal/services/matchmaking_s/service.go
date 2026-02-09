package matchmaking_s

import (
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

type QueuedPlayer struct {
	PlayerID pgtype.UUID
	Username string
}

type MatchmakingService struct {
	queue []QueuedPlayer
	mu    sync.Mutex
}

func NewMatchmakingService() *MatchmakingService {
	return &MatchmakingService{
		queue: make([]QueuedPlayer, 0),
	}
}

// EnterQueue tries to match the player with someone in the queue.
// If no one is waiting, adds the player to the queue.
// Returns the matched opponent if found, or nil if added to queue.
func (m *MatchmakingService) EnterQueue(playerID pgtype.UUID, username string) *QueuedPlayer {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if there's someone waiting in the queue
	if len(m.queue) > 0 {
		opponent := m.queue[0]
		m.queue = m.queue[1:]

		log.Info().
			Str("player1_id", opponent.PlayerID.String()).
			Str("player1_username", opponent.Username).
			Str("player2_id", playerID.String()).
			Str("player2_username", username).
			Msg("Match found! Two players matched")

		return &opponent
	}

	// If no one found add it to the queue
	m.queue = append(m.queue, QueuedPlayer{
		PlayerID: playerID,
		Username: username,
	})

	log.Info().
		Str("player_id", playerID.String()).
		Str("username", username).
		Int("queue_size", len(m.queue)).
		Msg("Player entered matchmaking queue")

	return nil
}

func (m *MatchmakingService) RemoveFromQueue(playerID pgtype.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, player := range m.queue {
		if player.PlayerID == playerID {
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			log.Info().
				Str("player_id", playerID.String()).
				Str("username", player.Username).
				Msg("Player removed from matchmaking queue")
			return
		}
	}
}

func (m *MatchmakingService) GetQueueSize() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.queue)
}
