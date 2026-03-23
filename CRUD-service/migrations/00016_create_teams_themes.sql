-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS teams_themes (
    id SERIAL PRIMARY KEY,
    team_id INTEGER NOT NULL,
    theme_id INTEGER NOT NULL,
    subtheme_id INTEGER,
    for_vip BOOLEAN NOT NULL,
    UNIQUE (theme_id, subtheme_id, for_vip)
);

ALTER TABLE teams_themes DROP CONSTRAINT IF EXISTS fk_teams_themes_team;
ALTER TABLE teams_themes ADD CONSTRAINT fk_teams_themes_team FOREIGN KEY (team_id) REFERENCES teams(id);

ALTER TABLE teams_themes DROP CONSTRAINT IF EXISTS fk_teams_themes_theme;
ALTER TABLE teams_themes ADD CONSTRAINT fk_teams_themes_theme FOREIGN KEY (theme_id) REFERENCES themes(id);

ALTER TABLE teams_themes DROP CONSTRAINT IF EXISTS fk_teams_themes_subtheme;
ALTER TABLE teams_themes ADD CONSTRAINT fk_teams_themes_subtheme FOREIGN KEY (subtheme_id) REFERENCES subthemes(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP CONSTRAINT IF EXISTS fk_teams_themes_theme;
DROP CONSTRAINT IF EXISTS fk_teams_themes_team;
DROP TABLE IF EXISTS teams_themes;
-- +goose StatementEnd
