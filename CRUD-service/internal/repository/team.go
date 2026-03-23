package repository

import (
	"fmt"
	"slices"

	"github.com/jmoiron/sqlx"

	"crud-service/internal/model"
)

type teamDB struct {
	ID            int    `db:"id"`
	Name          string `db:"name"`
	ThemeSubtheme []model.ThemeSubtheme
}

func toTeamDB(t model.Team) teamDB {
	return teamDB{ID: t.ID, Name: t.Name, ThemeSubtheme: t.ThemeSubtheme}
}

func (t teamDB) toDomain() model.Team {
	result := model.Team{ID: t.ID, Name: t.Name, ThemeSubtheme: t.ThemeSubtheme}
	return result
}

type TeamRepository interface {
	GetAll() ([]model.Team, error)
	GetByID(id int) (model.Team, error)
	Create(t model.Team) (model.Team, error)
	Update(id int, t model.Team) (model.Team, error)
	Delete(id int) error
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
	result := make([]model.Team, len(rows))
	for i, row := range rows {
		filledRow, err := r.fillThemesSubthemes(row)
		if err != nil {
			return nil, fmt.Errorf("teamRepo.GetAll getThemesSubthemes: %w", err)
		}

		result[i] = filledRow.toDomain()
	}
	return result, nil
}

func (r *teamRepo) GetByID(id int) (model.Team, error) {
	var row teamDB
	if err := r.db.Get(&row, `SELECT id, name FROM teams WHERE id = $1`, id); err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.GetByID: %w", wrapNotFound(err))
	}

	filledRow, err := r.fillThemesSubthemes(row)
	if err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.GetByID getThemesSubthemes: %w", err)
	}

	return filledRow.toDomain(), nil
}

func (r *teamRepo) Create(t model.Team) (model.Team, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.Create start transaction: %w", err)
	}

	teamInfo := toTeamDB(t)
	err = tx.QueryRowx(
		`INSERT INTO teams (name) VALUES ($1) RETURNING id`,
		teamInfo.Name,
	).Scan(&teamInfo.ID)
	if err != nil {
		rberr := tx.Rollback()
		return model.Team{}, fmt.Errorf("teamRepo.Create teams: %w, rollback error: %w", err, rberr)
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
			rberr := tx.Rollback()
			return model.Team{}, fmt.Errorf("teamRepo.Create teams_themes: %w, rollback error: %w", err, rberr)
		}
	}

	err = tx.Commit()
	if err != nil {
		rberr := tx.Rollback()
		return model.Team{}, fmt.Errorf("teamRepo.Create commit transaction: %w, rollback error: %w", err, rberr)
	}

	filledRow, err := r.fillThemesSubthemes(teamInfo)
	if err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.Create fillThemesSubthemes: %w", err)
	}

	return filledRow.toDomain(), nil
}

func (r *teamRepo) Update(id int, t model.Team) (model.Team, error) {
	teamInfo := toTeamDB(t)

	tx, err := r.db.Beginx()
	if err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.Update start transaction: %w", err)
	}

	_, err = tx.Exec(
		`UPDATE teams SET name=$1 WHERE id=$2`,
		teamInfo.Name,
		teamInfo.ID,
	)
	if err != nil {
		rberr := tx.Rollback()
		return model.Team{}, fmt.Errorf("teamRepo.Update teams: %w, rollback error: %w", err, rberr)
	}

	var currentThemes []model.ThemeSubtheme
	err = tx.Select(&currentThemes, `SELECT theme_id, subtheme_id FROM teams_themes WHERE team_id = $1`, teamInfo.ID)
	if err != nil {
		rberr := tx.Rollback()
		return model.Team{}, fmt.Errorf("teamRepo.Update teams_themes: %w, rollback error: %w", err, rberr)
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
			rberr := tx.Rollback()
			return model.Team{}, fmt.Errorf("teamRepo.Update teams_themes: %w, rollback error: %w", err, rberr)
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
				rberr := tx.Rollback()
				return model.Team{}, fmt.Errorf("teamRepo.Update teams_themes: %w, rollback error: %w", err, rberr)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		rberr := tx.Rollback()
		return model.Team{}, fmt.Errorf("teamRepo.Update commit transaction: %w, rollback error: %w", err, rberr)
	}

	filledRow, err := r.fillThemesSubthemes(teamInfo)
	if err != nil {
		return model.Team{}, fmt.Errorf("teamRepo.Update fillThemesSubthemes: %w", err)
	}

	return filledRow.toDomain(), nil
}

func (r *teamRepo) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM teams WHERE id=$1`, id)
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
		return model.Team{}, fmt.Errorf("teamRepo.GetTeamByThemeSubtheme: %w", wrapNotFound(err))
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
		return model.Team{}, fmt.Errorf("teamRepo.GetTeamByName: %w", wrapNotFound(err))
	}

	return team.toDomain(), nil
}
