-- VGC Champions init script (Postgres / Neon)
-- Order respects FKs. Source: docs/relational-model.md

CREATE SCHEMA IF NOT EXISTS vgcalc;

CREATE TABLE vgcalc.types (
  id           SERIAL PRIMARY KEY,
  pokeapi_id   INTEGER NOT NULL UNIQUE,
  slug         TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL
);

CREATE TABLE vgcalc.type_efficacy (
  attack_type_id INTEGER NOT NULL REFERENCES vgcalc.types (id),
  target_type_id INTEGER NOT NULL REFERENCES vgcalc.types (id),
  multiplier     NUMERIC(3, 1) NOT NULL CHECK (multiplier IN (0, 0.5, 1, 2)),
  PRIMARY KEY (attack_type_id, target_type_id)
);

CREATE TABLE vgcalc.abilities (
  id           SERIAL PRIMARY KEY,
  pokeapi_id   INTEGER NOT NULL UNIQUE,
  slug         TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  short_effect TEXT,
  effect       TEXT
);

CREATE TABLE vgcalc.moves (
  id           SERIAL PRIMARY KEY,
  pokeapi_id   INTEGER NOT NULL UNIQUE,
  slug         TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  short_effect TEXT,
  effect       TEXT,
  type_id      INTEGER NOT NULL REFERENCES vgcalc.types (id),
  category     TEXT NOT NULL CHECK (category IN ('physical', 'special', 'status')),
  power        INTEGER CHECK (power > 0),
  accuracy     INTEGER CHECK (accuracy BETWEEN 1 AND 100),
  pp           INTEGER NOT NULL CHECK (pp >= 1),
  priority     INTEGER NOT NULL DEFAULT 0,
  target       TEXT NOT NULL,
  is_contact   BOOLEAN NOT NULL DEFAULT FALSE,
  is_punch     BOOLEAN NOT NULL DEFAULT FALSE,
  is_bite      BOOLEAN NOT NULL DEFAULT FALSE,
  is_sound     BOOLEAN NOT NULL DEFAULT FALSE,
  is_pulse     BOOLEAN NOT NULL DEFAULT FALSE,
  is_bullet    BOOLEAN NOT NULL DEFAULT FALSE,
  is_slicing   BOOLEAN NOT NULL DEFAULT FALSE,
  is_wind      BOOLEAN NOT NULL DEFAULT FALSE,
  min_hits     INTEGER CHECK (min_hits >= 1),
  max_hits     INTEGER CHECK (max_hits >= 1),
  CHECK (max_hits IS NULL OR min_hits IS NULL OR max_hits >= min_hits)
);

CREATE TABLE vgcalc.items (
  id           SERIAL PRIMARY KEY,
  pokeapi_id   INTEGER NOT NULL UNIQUE,
  slug         TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  short_effect TEXT,
  effect       TEXT,
  fling_power  INTEGER CHECK (fling_power >= 0)
);

CREATE TABLE vgcalc.natures (
  id           SERIAL PRIMARY KEY,
  pokeapi_id   INTEGER NOT NULL UNIQUE,
  slug         TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  plus_stat    TEXT NOT NULL CHECK (plus_stat IN ('atk', 'def', 'spa', 'spd', 'spe')),
  minus_stat   TEXT NOT NULL CHECK (minus_stat IN ('atk', 'def', 'spa', 'spd', 'spe'))
);

CREATE TABLE vgcalc.pokemon (
  id                SERIAL PRIMARY KEY,
  pokeapi_id        INTEGER NOT NULL UNIQUE,
  slug              TEXT NOT NULL UNIQUE,
  display_name      TEXT NOT NULL,
  primary_type_id   INTEGER NOT NULL REFERENCES vgcalc.types (id),
  secondary_type_id INTEGER REFERENCES vgcalc.types (id),
  hp                INTEGER NOT NULL CHECK (hp >= 1),
  atk               INTEGER NOT NULL CHECK (atk >= 1),
  def               INTEGER NOT NULL CHECK (def >= 1),
  spa               INTEGER NOT NULL CHECK (spa >= 1),
  spd               INTEGER NOT NULL CHECK (spd >= 1),
  spe               INTEGER NOT NULL CHECK (spe >= 1),
  ability1_id       INTEGER NOT NULL REFERENCES vgcalc.abilities (id),
  ability2_id       INTEGER REFERENCES vgcalc.abilities (id),
  hidden_ability_id INTEGER REFERENCES vgcalc.abilities (id),
  weight            NUMERIC(6, 2) NOT NULL CHECK (weight > 0),
  fixed_gender      TEXT CHECK (fixed_gender IN ('male', 'female', 'genderless'))
);

CREATE TABLE vgcalc.mega_evolution (
  id                SERIAL PRIMARY KEY,
  pokeapi_id        INTEGER NOT NULL UNIQUE,
  base_pokemon_id   INTEGER NOT NULL REFERENCES vgcalc.pokemon (id),
  slug              TEXT NOT NULL UNIQUE,
  required_item_id  INTEGER REFERENCES vgcalc.items (id),
  required_move_id  INTEGER REFERENCES vgcalc.moves (id),
  ability_id        INTEGER NOT NULL REFERENCES vgcalc.abilities (id),
  primary_type_id   INTEGER NOT NULL REFERENCES vgcalc.types (id),
  secondary_type_id INTEGER REFERENCES vgcalc.types (id),
  hp                INTEGER NOT NULL CHECK (hp >= 1),
  atk               INTEGER NOT NULL CHECK (atk >= 1),
  def               INTEGER NOT NULL CHECK (def >= 1),
  spa               INTEGER NOT NULL CHECK (spa >= 1),
  spd               INTEGER NOT NULL CHECK (spd >= 1),
  spe               INTEGER NOT NULL CHECK (spe >= 1),
  weight            NUMERIC(6, 2) NOT NULL CHECK (weight > 0),
  CHECK (
    (CASE WHEN required_item_id IS NOT NULL THEN 1 ELSE 0 END +
     CASE WHEN required_move_id IS NOT NULL THEN 1 ELSE 0 END) = 1
  )
);

CREATE TABLE vgcalc.pokemon_moves (
  pokemon_id INTEGER NOT NULL REFERENCES vgcalc.pokemon (id) ON DELETE CASCADE,
  move_id    INTEGER NOT NULL REFERENCES vgcalc.moves (id) ON DELETE CASCADE,
  PRIMARY KEY (pokemon_id, move_id)
);

CREATE TABLE vgcalc.regulations (
  id           SERIAL PRIMARY KEY,
  slug         TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  is_current   BOOLEAN NOT NULL DEFAULT FALSE
);

-- Exactly one current regulation (allows zero, blocks two TRUEs).
CREATE UNIQUE INDEX regulations_single_current
  ON vgcalc.regulations (is_current) WHERE is_current = TRUE;

CREATE TABLE vgcalc.regulation_pokemon (
  regulation_id INTEGER NOT NULL REFERENCES vgcalc.regulations (id) ON DELETE CASCADE,
  pokemon_id    INTEGER NOT NULL REFERENCES vgcalc.pokemon (id) ON DELETE CASCADE,
  PRIMARY KEY (regulation_id, pokemon_id)
);

-- FK lookup indexes
CREATE INDEX idx_pokemon_primary_type ON vgcalc.pokemon (primary_type_id);
CREATE INDEX idx_pokemon_ability1 ON vgcalc.pokemon (ability1_id);
CREATE INDEX idx_mega_base ON vgcalc.mega_evolution (base_pokemon_id);
CREATE INDEX idx_moves_type ON vgcalc.moves (type_id);
CREATE INDEX idx_pokemon_moves_move ON vgcalc.pokemon_moves (move_id);
CREATE INDEX idx_regulation_pokemon_pokemon ON vgcalc.regulation_pokemon (pokemon_id);
