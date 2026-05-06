-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS teams_themes (
    id SERIAL PRIMARY KEY,
    team_id INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    theme_id INTEGER NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
    subtheme_id INTEGER REFERENCES subthemes(id) ON DELETE CASCADE,
    for_vip BOOLEAN NOT NULL,
    UNIQUE NULLS NOT DISTINCT (theme_id, subtheme_id, for_vip)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS teams_themes;
-- +goose StatementEnd
