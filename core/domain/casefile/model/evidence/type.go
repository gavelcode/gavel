package evidence

import "fmt"

type Type struct {
	value string
}

var (
	TypeSourceCode   = Type{value: "source_code"}
	TypeSecurity     = Type{value: "security"}
	TypeSupplyChain  = Type{value: "supply_chain"}
	TypeCoverage     = Type{value: "coverage"}
	TypeArchitecture = Type{value: "architecture"}
)

var validTypes = map[string]Type{
	"source_code":  TypeSourceCode,
	"security":     TypeSecurity,
	"supply_chain": TypeSupplyChain,
	"coverage":     TypeCoverage,
	"architecture": TypeArchitecture,
}

func NewType(s string) (Type, error) {
	typ, ok := validTypes[s]
	if !ok {
		return Type{}, fmt.Errorf("%w: %q", ErrInvalidType, s)
	}
	return typ, nil
}

func (t Type) String() string {
	return t.value
}

func (t Type) Equal(other Type) bool {
	return t.value == other.value
}
