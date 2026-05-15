package updatelanguages

import (
	"fmt"
	"strings"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/coverage"
)

type Command struct {
	projectID string
	languages []coverage.Language
}

func NewCommand(projectID string, languages []string) (Command, error) {
	if strings.TrimSpace(projectID) == "" {
		return Command{}, fmt.Errorf("%w: projectID must not be empty", ErrInvalidCommand)
	}
	parsed, err := parseLanguages(languages)
	if err != nil {
		return Command{}, err
	}
	return Command{
		projectID: projectID,
		languages: parsed,
	}, nil
}

func parseLanguages(names []string) ([]coverage.Language, error) {
	parsed := make([]coverage.Language, 0, len(names))
	for i, name := range names {
		lang, err := coverage.NewLanguage(name)
		if err != nil {
			return nil, fmt.Errorf("%w: languages[%d]: %s", ErrInvalidCommand, i, err.Error())
		}
		parsed = append(parsed, lang)
	}
	return parsed, nil
}

func (c Command) ProjectID() string { return c.projectID }

func (c Command) Languages() []coverage.Language {
	copied := make([]coverage.Language, len(c.languages))
	copy(copied, c.languages)
	return copied
}
