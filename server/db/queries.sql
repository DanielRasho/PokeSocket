
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