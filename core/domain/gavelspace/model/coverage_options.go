package model

type CoverageOptions struct {
	testSizeFilters       string
	testTagFilters        string
	instrumentationFilter string
}

func NewCoverageOptions(testSizeFilters, testTagFilters, instrumentationFilter string) CoverageOptions {
	return CoverageOptions{
		testSizeFilters:       testSizeFilters,
		testTagFilters:        testTagFilters,
		instrumentationFilter: instrumentationFilter,
	}
}

func (c CoverageOptions) TestSizeFilters() string       { return c.testSizeFilters }
func (c CoverageOptions) TestTagFilters() string        { return c.testTagFilters }
func (c CoverageOptions) InstrumentationFilter() string { return c.instrumentationFilter }
