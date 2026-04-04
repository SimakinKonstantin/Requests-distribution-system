package workflow

// TeamAssigner abstracts team assignment to avoid importing the CRUD service package here.
type TeamAssigner interface {
	AssignTeam(appealId int, teamId int) error
}
