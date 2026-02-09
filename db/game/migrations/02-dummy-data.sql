
INSERT INTO moves (name, type, power, accuracy, pp, effect_description) VALUES
-- Normal type moves
('Tackle', 'normal', 40, 100, 35, 'A physical attack in which the user charges and slams into the target.'),
('Quick Attack', 'normal', 40, 100, 30, 'An almost-invisibly fast attack that always strikes first.'),
('Hyper Beam', 'normal', 150, 90, 5, 'Powerful attack but user must rest next turn.'),
('Body Slam', 'normal', 85, 100, 15, 'May paralyze the target.'),

-- Fire type moves
('Ember', 'fire', 40, 100, 25, 'A weak fire attack that may inflict a burn.'),
('Flamethrower', 'fire', 90, 100, 15, 'A powerful fire attack that may inflict a burn.'),
('Fire Blast', 'fire', 110, 85, 5, 'The most powerful fire attack with burn chance.'),

-- Water type moves
('Water Gun', 'water', 40, 100, 25, 'Squirts water to attack the target.'),
('Surf', 'water', 90, 100, 15, 'A big wave crashes down on the target.'),
('Hydro Pump', 'water', 110, 80, 5, 'Blasts water at high power to strike the target.'),

-- Grass type moves
('Vine Whip', 'grass', 45, 100, 25, 'Strikes with slender vines.'),
('Razor Leaf', 'grass', 55, 95, 25, 'Sharp-edged leaves are launched to slash.'),
('Solar Beam', 'grass', 120, 100, 10, 'Absorbs light then blasts a beam.'),

-- Electric type moves
('Thunder Shock', 'electric', 40, 100, 30, 'A jolt of electricity crashes down.'),
('Thunderbolt', 'electric', 90, 100, 15, 'A strong blast of electricity.'),
('Thunder', 'electric', 110, 70, 10, 'A wicked thunderbolt that may paralyze.'),

-- Psychic type moves
('Confusion', 'psychic', 50, 100, 25, 'A weak telekinetic attack.'),
('Psychic', 'psychic', 90, 100, 10, 'Strong telekinetic attack.'),

-- Ice type moves
('Ice Beam', 'ice', 90, 100, 10, 'Fires an icy cold beam that may freeze.'),
('Blizzard', 'ice', 110, 70, 5, 'A howling blizzard that may freeze.'),

-- Fighting type moves
('Karate Chop', 'fighting', 50, 100, 25, 'A sharp chop with high critical hit ratio.'),
('Low Kick', 'fighting', 65, 100, 20, 'A powerful low kick attack.'),

-- Flying type moves
('Peck', 'flying', 35, 100, 35, 'A basic flying-type attack.'),
('Wing Attack', 'flying', 60, 100, 35, 'Strikes with wings spread wide.'),

-- Poison type moves
('Poison Sting', 'poison', 15, 100, 35, 'May poison the target.'),
('Sludge Bomb', 'poison', 90, 100, 10, 'Hurls sludge that may poison.');

INSERT INTO pokemon_species (name, base_hp, base_attack, base_defense, base_speed, type1, type2, sprite_url) VALUES
('Charizard', 78, 84, 78, 100, 'fire', 'flying', 'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/6.png'),
('Blastoise', 79, 83, 100, 78, 'water', NULL, 'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/9.png'),
('Venusaur', 80, 82, 83, 80, 'grass', 'poison', 'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/3.png'),
('Pikachu', 35, 55, 40, 90, 'electric', NULL, 'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/25.png'),
('Gengar', 60, 65, 60, 110, 'ghost', 'poison', 'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/94.png'),
('Alakazam', 55, 50, 45, 120, 'psychic', NULL, 'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/65.png'),
('Machamp', 90, 130, 80, 55, 'fighting', NULL, 'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/68.png'),
('Gyarados', 95, 125, 79, 81, 'water', 'flying', 'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/130.png'),
('Dragonite', 91, 134, 95, 80, 'dragon', 'flying', 'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/149.png'),
('Lapras', 130, 85, 80, 60, 'water', 'ice', 'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/131.png');

-- Charizard (Fire/Flying)
INSERT INTO pokemon_moves (pokemon_species_id, move_id) VALUES
(1, 6),  -- Flamethrower
(1, 7),  -- Fire Blast
(1, 23), -- Wing Attack
(1, 4);  -- Body Slam

-- Blastoise (Water)
INSERT INTO pokemon_moves (pokemon_species_id, move_id) VALUES
(2, 9),  -- Surf
(2, 10), -- Hydro Pump
(2, 19), -- Ice Beam
(2, 4);  -- Body Slam

-- Venusaur (Grass/Poison)
INSERT INTO pokemon_moves (pokemon_species_id, move_id) VALUES
(3, 11), -- Vine Whip
(3, 12), -- Razor Leaf
(3, 13), -- Solar Beam
(3, 25); -- Sludge Bomb

-- Pikachu (Electric)
INSERT INTO pokemon_moves (pokemon_species_id, move_id) VALUES
(4, 14), -- Thunder Shock
(4, 15), -- Thunderbolt
(4, 16), -- Thunder
(4, 2);  -- Quick Attack

-- Gengar (Ghost/Poison)
INSERT INTO pokemon_moves (pokemon_species_id, move_id) VALUES
(5, 17), -- Confusion
(5, 18), -- Psychic
(5, 25), -- Sludge Bomb
(5, 6);  -- Flamethrower

-- Alakazam (Psychic)
INSERT INTO pokemon_moves (pokemon_species_id, move_id) VALUES
(6, 17), -- Confusion
(6, 18), -- Psychic
(6, 15), -- Thunderbolt
(6, 2);  -- Quick Attack

-- Machamp (Fighting)
INSERT INTO pokemon_moves (pokemon_species_id, move_id) VALUES
(7, 21), -- Karate Chop
(7, 22), -- Low Kick
(7, 1),  -- Tackle
(7, 4);  -- Body Slam

-- Gyarados (Water/Flying)
INSERT INTO pokemon_moves (pokemon_species_id, move_id) VALUES
(8, 9),  -- Surf
(8, 10), -- Hydro Pump
(8, 23), -- Wing Attack
(8, 3);  -- Hyper Beam

-- Dragonite (Dragon/Flying)
INSERT INTO pokemon_moves (pokemon_species_id, move_id) VALUES
(9, 23), -- Wing Attack
(9, 6),  -- Flamethrower
(9, 3),  -- Hyper Beam
(9, 15); -- Thunderbolt

-- Lapras (Water/Ice)
INSERT INTO pokemon_moves (pokemon_species_id, move_id) VALUES
(10, 9),  -- Surf
(10, 19), -- Ice Beam
(10, 20), -- Blizzard
(10, 4);  -- Body Slam