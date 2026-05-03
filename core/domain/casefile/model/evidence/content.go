package evidence

type Content interface {
	Type() Type
	Subtype() Subtype
	Merge(other Content) (Content, error)
}
