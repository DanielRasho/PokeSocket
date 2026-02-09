CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE pokemon_species (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    base_hp INTEGER NOT NULL,
    base_attack INTEGER NOT NULL,
    base_defense INTEGER NOT NULL,
    base_speed INTEGER NOT NULL,
    type1 VARCHAR(20) NOT NULL, -- e.g., 'fire', 'water', 'grass'
    type2 VARCHAR(20), -- nullable for single-type Pokemon
    sprite_url VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE moves (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    type VARCHAR(20) NOT NULL,
    power INTEGER NOT NULL,
    accuracy INTEGER NOT NULL, -- 0-100
    pp INTEGER NOT NULL, -- Power Points (how many times it can be used)
    effect_description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pokemon_moves (
    pokemon_species_id INTEGER REFERENCES pokemon_species(id) ON DELETE CASCADE,
    move_id INTEGER REFERENCES moves(id) ON DELETE CASCADE,
    PRIMARY KEY (pokemon_species_id, move_id)
);

-- ============================================
-- USER & SESSION MANAGEMENT
-- ============================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) NOT NULL,
    connected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'connected' -- 'connected', 'in_matchmaking', 'in_battle', 'disconnected'
);

CREATE TABLE user_team (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    pokemon_species_id INTEGER REFERENCES pokemon_species(id),
    position INTEGER NOT NULL CHECK (position >= 1 AND position <= 6), -- slot in team (1-6)
    current_hp INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT FALSE, -- currently in battle
    is_fainted BOOLEAN DEFAULT FALSE,
    UNIQUE(user_id, position)
);

-- ============================================
-- BATTLES
-- ============================================

-- Stores battle instances
CREATE TABLE battles (
    id UUID PRIMARY KEY,
    player1_id UUID REFERENCES users(id) ON DELETE SET NULL,
    player2_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'completed', 'abandoned'
    winner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    current_turn INTEGER DEFAULT 1,
    player1_active_pokemon_position INTEGER, -- which slot is currently active
    player2_active_pokemon_position INTEGER,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    
    CHECK (player1_id != player2_id)
);

-- Stores battle state for each player's team during battle
CREATE TABLE battle_pokemon (
    id SERIAL PRIMARY KEY,
    battle_id UUID REFERENCES battles(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    pokemon_species_id INTEGER REFERENCES pokemon_species(id),
    position INTEGER NOT NULL CHECK (position >= 1 AND position <= 6),
    current_hp INTEGER NOT NULL,
    is_fainted BOOLEAN DEFAULT FALSE,
    
    UNIQUE(battle_id, user_id, position)
);

CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_battles_status ON battles(status);
CREATE INDEX idx_battles_players ON battles(player1_id, player2_id);