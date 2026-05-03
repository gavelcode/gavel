package evidence

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidLanguage      = failure.New("invalid language", failure.Validation)
	ErrInvalidLanguageStats = failure.New("invalid language coverage", failure.Validation)
	ErrInvalidDependency    = failure.New("invalid dependency license", failure.Validation)
	ErrInvalidType          = failure.New("invalid evidence type", failure.Validation)
	ErrInvalidSubtype       = failure.New("invalid evidence subtype", failure.Validation)
	ErrInvalidContent       = failure.New("invalid evidence content", failure.Validation)
	ErrInvalidEvidence      = failure.New("invalid evidence", failure.Validation)
	ErrInvalidViolation     = failure.New("invalid architecture violation", failure.Validation)
)
