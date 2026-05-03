package verdict

import "github.com/usegavel/gavel/core/domain/casefile/model/evidence"

type Ruling struct {
	subtype evidence.Subtype
	passed  bool
	detail  string
}

func NewRuling(subtype evidence.Subtype, passed bool, detail string) Ruling {
	return Ruling{
		subtype: subtype,
		passed:  passed,
		detail:  detail,
	}
}

func (r Ruling) Subtype() evidence.Subtype { return r.subtype }
func (r Ruling) Passed() bool              { return r.passed }
func (r Ruling) Detail() string            { return r.detail }
