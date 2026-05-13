package createcasefile

type Option func(*Command)

func WithFreshEvaluation() Option {
	return func(c *Command) { c.freshEvaluation = true }
}
