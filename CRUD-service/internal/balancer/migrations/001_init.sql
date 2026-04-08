-- Demo schema (minimal)

CREATE TABLE IF NOT EXISTS appeals (
  id                          integer PRIMARY KEY,
  team_id                     text NOT NULL,
  is_urgent                    boolean NOT NULL DEFAULT false,
  is_important                 boolean NOT NULL DEFAULT false,
  created_at                   timestamptz NOT NULL DEFAULT now(),
  pending_client_message_created_at timestamptz NULL,
  manager_id                   text NULL,
  status                       text NOT NULL DEFAULT 'new'
);

CREATE TABLE IF NOT EXISTS managers (
  id                   text PRIMARY KEY,
  is_available          boolean NOT NULL DEFAULT true,
  active_appeals_count  integer NOT NULL DEFAULT 0,
  last_assign_at        timestamptz NULL
);

CREATE TABLE IF NOT EXISTS manager_teams (
  manager_id text NOT NULL REFERENCES managers(id) ON DELETE CASCADE,
  team_id    text NOT NULL,
  PRIMARY KEY (manager_id, team_id)
);

CREATE TABLE IF NOT EXISTS slots (
  id         text PRIMARY KEY,
  manager_id text NOT NULL REFERENCES managers(id) ON DELETE CASCADE,
  appeal_id  integer NULL REFERENCES appeals(id) ON DELETE SET NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Postgres "sorted set" analogue
CREATE TABLE IF NOT EXISTS pending_appeals (
  appeal_id  integer PRIMARY KEY REFERENCES appeals(id) ON DELETE CASCADE,
  team_id    text NOT NULL,
  priority   double precision NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS pending_appeals_priority_idx ON pending_appeals (priority DESC, updated_at ASC);
CREATE INDEX IF NOT EXISTS slots_free_idx ON slots (manager_id, updated_at) WHERE appeal_id IS NULL;

