package evidence

import "fmt"

type Subtype struct {
	value      string
	parentType Type
}

var (
	SubtypeCodeQuality     = Subtype{value: "code_quality", parentType: TypeSourceCode}
	SubtypeComplexity      = Subtype{value: "complexity", parentType: TypeSourceCode}
	SubtypeSAST            = Subtype{value: "sast", parentType: TypeSecurity}
	SubtypeSecrets         = Subtype{value: "secrets", parentType: TypeSecurity}
	SubtypeMalware         = Subtype{value: "malware", parentType: TypeSecurity}
	SubtypeDAST            = Subtype{value: "dast", parentType: TypeSecurity}
	SubtypeSCA             = Subtype{value: "sca", parentType: TypeSupplyChain}
	SubtypeLicense         = Subtype{value: "license", parentType: TypeSupplyChain}
	SubtypeCoverage        = Subtype{value: "coverage", parentType: TypeCoverage}
	SubtypeNewCodeCoverage = Subtype{value: "new_code_coverage", parentType: TypeCoverage}
	SubtypeArchitecture    = Subtype{value: "architecture", parentType: TypeArchitecture}
)

var validSubtypes = map[string]Subtype{
	"code_quality":      SubtypeCodeQuality,
	"complexity":        SubtypeComplexity,
	"sast":              SubtypeSAST,
	"secrets":           SubtypeSecrets,
	"malware":           SubtypeMalware,
	"dast":              SubtypeDAST,
	"sca":               SubtypeSCA,
	"license":           SubtypeLicense,
	"coverage":          SubtypeCoverage,
	"new_code_coverage": SubtypeNewCodeCoverage,
	"architecture":      SubtypeArchitecture,
}

var nonFindingBasedSubtypes = map[Subtype]bool{
	SubtypeCoverage:        true,
	SubtypeNewCodeCoverage: true,
	SubtypeLicense:         true,
	SubtypeArchitecture:    true,
}

func NewSubtype(s string) (Subtype, error) {
	sub, ok := validSubtypes[s]
	if !ok {
		return Subtype{}, fmt.Errorf("%w: %q", ErrInvalidSubtype, s)
	}
	return sub, nil
}

func (s Subtype) String() string {
	return s.value
}

func (s Subtype) Type() Type {
	return s.parentType
}

func (s Subtype) Equal(other Subtype) bool {
	return s.value == other.value && s.parentType.Equal(other.parentType)
}

func IsSubtypeFindingBased(subtype Subtype) bool {
	return !nonFindingBasedSubtypes[subtype]
}
