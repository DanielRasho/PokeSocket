
-- name: InsertUser :one
INSERT INTO users (id, username, status)
VALUES (@id, @username, 'connected')
RETURNING id;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = @id;

-- name: InsertUserTeamPokemon :exec
INSERT INTO user_team (user_id, pokemon_species_id, position, current_hp, is_active, is_fainted)
SELECT @user_id, @pokemon_species_id, @position, ps.base_hp, false, false
FROM pokemon_species ps
WHERE ps.id = @pokemon_species_id;

-- name: GetPokemonSpecies :one
SELECT id, name, base_hp, base_attack, base_defense, base_speed, type1, type2
FROM pokemon_species
WHERE id = @id;

-- name: CreateBattle :one
INSERT INTO battles (id, player1_id, player2_id, status, player1_active_pokemon_position, player2_active_pokemon_position)
VALUES (@id, @player1_id, @player2_id, 'active', 1, 1)
RETURNING id, player1_id, player2_id, status, started_at;

-- name: DeleteBattle :exec
DELETE FROM battles
WHERE id = @id;

-- name: GetUserTeam :many
SELECT id, user_id, pokemon_species_id, position, current_hp, is_active, is_fainted
FROM user_team
WHERE user_id = @user_id
ORDER BY position;

-- name: GetBattle :one
SELECT id, player1_id, player2_id, status, current_turn, player1_active_pokemon_position, player2_active_pokemon_position
FROM battles
WHERE id = @id;

-- name: UpdatePokemonHP :exec
UPDATE user_team
SET current_hp = @current_hp,
    is_fainted = CASE WHEN @current_hp <= 0 THEN true ELSE is_fainted END
WHERE user_id = @user_id AND position = @position;

-- name: UpdateBattleTurn :exec
UPDATE battles
SET current_turn = current_turn + 1
WHERE id = @id;

-- name: UpdatePlayer1ActivePokemon :exec
UPDATE battles
SET player1_active_pokemon_position = @player1_active_pokemon_position
WHERE id = @id;

-- name: UpdatePlayer2ActivePokemon :exec
UPDATE battles
SET player2_active_pokemon_position = @player2_active_pokemon_position
WHERE id = @id;