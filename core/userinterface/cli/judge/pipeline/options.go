package pipeline

type Options struct {
	Quick            bool
	Absolute         bool
	NoBaselineUpdate bool
	RequireSubmit    bool
	PRNumber         int
	PRTitle          string
	PRAuthor         string
	PRBranch         string
	Gavelspace       string
	TargetPattern    string
	Workspace        string
}
