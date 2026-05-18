package gavelconfig

type CoverageOptions struct {
	testSizeFilters       string
	testTagFilters        string
	instrumentationFilter string
}

func (c CoverageOptions) TestSizeFilters() string       { return c.testSizeFilters }
func (c CoverageOptions) TestTagFilters() string        { return c.testTagFilters }
func (c CoverageOptions) InstrumentationFilter() string { return c.instrumentationFilter }
func (c CoverageOptions) IsZero() bool {
	return c.testSizeFilters == "" && c.testTagFilters == "" && c.instrumentationFilter == ""
}
