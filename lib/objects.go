package lib

type PositionCreate struct {
	Id       int32   `db:"id"`
	Role     string  `db:"role"`
	Start    string  `db:"start"`
	End      *string `db:"end"`
	WorkDone string  `db:"workdone"`
	Projects []int32 `db:"projects"`
}
