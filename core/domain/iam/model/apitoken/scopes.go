package apitoken

type Scopes []Scope

func (s Scopes) Contains(target Scope) bool {
	for _, sc := range s {
		if sc.Equal(target) {
			return true
		}
	}
	return false
}
