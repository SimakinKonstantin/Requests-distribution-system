package workflow

type TeamAssigner interface {
	AssignTeam(appealId int, teamId int) error
}
