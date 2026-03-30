package repository

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/jmoiron/sqlx"

	"crud-service/internal/model"
)

type ThemeSubthemeDB struct {
	ThemeID    int  `db:"theme_id"`
	SubthemeID int  `db:"subtheme_id"`
	ForVip     bool `db:"for_vip"`
}

type teamDB struct {
	ID            int    `db:"id"`
	Name          string `db:"name"`
	ThemeSubtheme []ThemeSubthemeDB
}

func toTeamDB(t model.Team) teamDB {
	result := teamDB{ID: t.ID, Name: t.Name}
	for _, themeSubtheme := range t.ThemeSubtheme {
		result.ThemeSubtheme = append(result.ThemeSubtheme, ThemeSubthemeDB{ThemeID: themeSubtheme.ThemeID, SubthemeID: themeSubtheme.SubthemeID, ForVip: themeSubtheme.ForVip})
	}
	return result
}

func (t teamDB) toDomain() model.Team {
	result := model.Team{ID: t.ID, Name: t.Name}
	for _, themeSubtheme := range t.ThemeSubtheme {
		result.ThemeSubtheme = append(result.ThemeSubtheme, model.ThemeSubtheme{ThemeID: themeSubtheme.ThemeID, SubthemeID: themeSubtheme.SubthemeID, ForVip: themeSubtheme.ForVip})
	}
	return result
}

type TeamRepository interface {
	GetAll() ([]model.Team, error)
	GetByID(id int) (model.Team, error)
	Create(tx *sqlx.Tx, t model.Team) (model.Team, error)
	Update(tx *sqlx.Tx, id int, t model.Team) (model.Team, error)
	Delete(tx *sqlx.Tx, id int) error
	GetTeamByThemeSubtheme(themeID int, subthemeID *int, isVIP bool) (model.Team, error)
	GetTeamByName(name string) (model.Team, error)
}

type teamRepo struct {
	db *sqlx.DB
}

// NewThemeRepository returns a PostgreSQL-backed ThemeRepository.
func NewTeamRepository(db *sqlx.DB) TeamRepository {
	return &teamRepo{db: db}
}

func (r *teamRepo) GetAll() ([]model.Team, error) {
	var rows []teamDB
	if err := r.db.Select(&rows, `SELECT id, name FROM teams ORDER BY id`); err != nil {
		return nil, fmt.Errorf("teamRepo.GetAll: %w", err)
	}

	slog.Warn("GOT TEAMS ROWS")

	result := make([]model.Team, len(rows))
	for i, row := range rows {
		filledRow, err := r.fillThemesSubthemes(row)
		if err != nil {
			return nil, fmt.Errorf("teamRepo.GetAll getThemesSubthemes: %w", err)
		}

		result[i] = filledRow.toDomain()
	}

	slog.Warn("GOT TEAMS SUBTHEMES")
	return result, nil
}

func (r *teamRepo) GetByID(id int) (model.Team, error) {
	var row teamDB
	if err := r.db.Get(&row, `SELECT id, name FROM teams WHERE id = $1`, id); err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.GetByID: %w", err)
	}

	filledRow, err := r.fillThemesSubthemes(row)
	if err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.GetByID getThemesSubthemes: %w", err)
	}

	return filledRow.toDomain(), nil
}

func (r *teamRepo) Create(tx *sqlx.Tx, t model.Team) (model.Team, error) {
	var err error
	teamInfo := toTeamDB(t)
	err = tx.QueryRowx(
		`INSERT INTO teams (name) VALUES ($1) RETURNING id`,
		teamInfo.Name,
	).Scan(&teamInfo.ID)
	if err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.Create teams: %w", err)
	}

	for _, themeSubtheme := range teamInfo.ThemeSubtheme {
		_, err = tx.Exec(
			`INSERT INTO teams_themes (team_id, theme_id, subtheme_id, for_vip) VALUES ($1, $2, $3, $4)`,
			teamInfo.ID,
			themeSubtheme.ThemeID,
			themeSubtheme.SubthemeID,
			themeSubtheme.ForVip,
		)
		if err != nil {
			return model.Team{}, fmt.Errorf("teamRepo.Create teams_themes: %w", err)
		}
	}

	filledRow, err := r.fillThemesSubthemes(teamInfo)
	if err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.Create fillThemesSubthemes: %w", err)
	}

	return filledRow.toDomain(), nil
}

func (r *teamRepo) Update(tx *sqlx.Tx, id int, t model.Team) (model.Team, error) {
	teamInfo := toTeamDB(t)
	var err error

	_, err = tx.Exec(
		`UPDATE teams SET name=$1 WHERE id=$2`,
		teamInfo.Name,
		teamInfo.ID,
	)
	if err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.Update teams: %w", err)
	}

	var currentThemes []ThemeSubthemeDB
	err = tx.Select(&currentThemes, `SELECT theme_id, subtheme_id FROM teams_themes WHERE team_id = $1`, teamInfo.ID)
	if err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.Update teams_themes: %w", err)
	}

	for _, themeSubtheme := range teamInfo.ThemeSubtheme {
		if !slices.Contains(currentThemes, themeSubtheme) {
			_, err = tx.Exec(
				`INSERT INTO teams_themes (team_id, theme_id, subtheme_id, for_vip) VALUES ($1, $2, $3, $4)`,
				teamInfo.ID,
				themeSubtheme.ThemeID,
				themeSubtheme.SubthemeID,
				themeSubtheme.ForVip,
			)
		}
		if err != nil {
			return model.Team{}, fmt.Errorf("teamRepo.Update teams_themes: %w", err)
		}
	}

	// Удаляем лишние темы
	for _, themeSubtheme := range currentThemes {
		if !slices.Contains(teamInfo.ThemeSubtheme, themeSubtheme) {
			_, err = tx.Exec(
				`DELETE FROM teams_themes WHERE team_id=$1 AND theme_id = $2 AND subtheme_id = $3`,
				teamInfo.ID,
				themeSubtheme.ThemeID,
				themeSubtheme.SubthemeID,
			)
			if err != nil {
				return model.Team{}, fmt.Errorf("teamRepo.Update teams_themes: %w", err)
			}
		}
	}

	filledRow, err := r.fillThemesSubthemes(teamInfo)
	if err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.Update fillThemesSubthemes: %w", err)
	}

	return filledRow.toDomain(), nil
}

func (r *teamRepo) Delete(tx *sqlx.Tx, id int) error {
	res, err := tx.Exec(`DELETE FROM teams WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("teamRepo.Delete: %w", err)
	}
	return expectOneRow(res)
}

func (r *teamRepo) fillThemesSubthemes(team teamDB) (teamDB, error) {
	result := team

	err := r.db.Select(&result.ThemeSubtheme, `SELECT theme_id, subtheme_id FROM teams_themes WHERE team_id = $1`, team.ID)
	if err != nil {
		return teamDB{}, fmt.Errorf("teamRepo.getThemesSubthemes teams_themes: %w", err)
	}

	return teamDB{ID: team.ID, Name: team.Name, ThemeSubtheme: result.ThemeSubtheme}, nil
}

func (r *teamRepo) GetTeamByThemeSubtheme(themeID int, subthemeID *int, isVIP bool) (model.Team, error) {
	var teamID int
	if err := r.db.Get(&teamID, `SELECT team_id FROM teams_themes WHERE theme_id = $1 AND subtheme_id = $2 AND for_vip = $3`, themeID, subthemeID, isVIP); err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.GetTeamByThemeSubtheme: %w", err)
	}

	team, err := r.GetByID(teamID)
	if err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.GetTeamByThemeSubtheme: %w", err)
	}

	return team, nil
}

func (r *teamRepo) GetTeamByName(name string) (model.Team, error) {
	var team teamDB
	if err := r.db.Get(&team, `SELECT id, name FROM teams WHERE name = $1`, name); err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.GetTeamByName: %w", err)
	}

	return team.toDomain(), nil
}
