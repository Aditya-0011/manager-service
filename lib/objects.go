package lib

type PositionCreate struct {
	Id       int32   `db:"id" json:"id"`
	Role     string  `db:"role" json:"role"`
	Start    string  `db:"start" json:"start"`
	End      *string `db:"end" json:"end"`
	WorkDone string  `db:"workdone" json:"workdone"`
	Projects []int32 `db:"projects" json:"projects"`
}

func (p PositionCreate) IsNull() bool { return false }
func (p PositionCreate) Index(i int) any {
	switch i {
	case 0:
		return p.Id
	case 1:
		return p.Role
	case 2:
		return p.Start
	case 3:
		return p.End
	case 4:
		return p.WorkDone
	case 5:
		return p.Projects
	default:
		return nil
	}
}
