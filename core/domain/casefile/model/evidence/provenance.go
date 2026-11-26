package evidence

import "fmt"

// Provenance says who stands behind a piece of evidence: gavel itself (a
// catalog tool, built hermetically with a guaranteed SARIF contract) or the
// project (a bridge it declared, which gavel only runs and validates). The
// zero value is catalog, so evidence that predates bridges keeps its meaning.
type Provenance struct {
	bridge bool
}

var (
	ProvenanceCatalog = Provenance{}
	ProvenanceBridge  = Provenance{bridge: true}
)

func NewProvenance(value string) (Provenance, error) {
	switch value {
	case "", "catalog":
		return ProvenanceCatalog, nil
	case "bridge":
		return ProvenanceBridge, nil
	default:
		return Provenance{}, fmt.Errorf("%w: %q", ErrInvalidProvenance, value)
	}
}

func (p Provenance) String() string {
	if p.bridge {
		return "bridge"
	}
	return "catalog"
}

func (p Provenance) IsBridge() bool { return p.bridge }

// GuaranteedByGavel is false for bridges: gavel does not vouch for tools the project provides.
func (p Provenance) GuaranteedByGavel() bool { return !p.bridge }
