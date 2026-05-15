package get

import "time"

type GavelspaceDetail struct {
	Name      string
	Projects  []ProjectRefView
	CreatedAt time.Time
}
