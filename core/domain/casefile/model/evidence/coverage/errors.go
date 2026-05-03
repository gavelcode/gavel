package coverage

import "github.com/usegavel/gavel/core/domain/shared/failure"

var (
	ErrInvalidLanguage      = failure.New("invalid language", failure.Validation)
	ErrInvalidLanguageStats = failure.New("invalid language coverage", failure.Validation)
	ErrInvalidContent       = failure.New("invalid coverage content", failure.Validation)
	ErrInvalidPatchContent  = failure.New("invalid new code coverage", failure.Validation)
)
