package evidence

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Evidence struct {
	id          EvidenceID
	subtype     Subtype
	source      string
	content     Content
	collectedAt time.Time
}

func NewEvidence(subtype Subtype, source string, content Content, collectedAt time.Time) (Evidence, error) {
	evidenceID := NewEvidenceID(uuid.New())
	return buildEvidence(evidenceID, subtype, source, content, collectedAt)
}

func ReconstituteEvidence(evidenceID EvidenceID, subtype Subtype, source string, content Content, collectedAt time.Time) (Evidence, error) {
	return buildEvidence(evidenceID, subtype, source, content, collectedAt)
}

func buildEvidence(evidenceID EvidenceID, subtype Subtype, source string, content Content, collectedAt time.Time) (Evidence, error) {
	if strings.TrimSpace(source) == "" {
		return Evidence{}, fmt.Errorf("%w: source must not be empty", ErrInvalidEvidence)
	}
	if content == nil {
		return Evidence{}, fmt.Errorf("%w: content must not be nil", ErrInvalidEvidence)
	}
	if collectedAt.IsZero() {
		return Evidence{}, fmt.Errorf("%w: collectedAt must not be zero", ErrInvalidEvidence)
	}
	if content.Subtype() != subtype {
		return Evidence{}, fmt.Errorf("%w: content subtype %q does not match expected subtype %q", ErrInvalidEvidence, content.Subtype(), subtype)
	}

	return Evidence{
		id:          evidenceID,
		subtype:     subtype,
		source:      source,
		content:     content,
		collectedAt: collectedAt,
	}, nil
}

func (e Evidence) ID() EvidenceID         { return e.id }
func (e Evidence) Subtype() Subtype       { return e.subtype }
func (e Evidence) Source() string         { return e.source }
func (e Evidence) Content() Content       { return e.content }
func (e Evidence) CollectedAt() time.Time { return e.collectedAt }
func (e Evidence) Type() Type             { return e.subtype.Type() }
