package gavelconfig

type coverageOptionsDTO struct {
	TestSizeFilters       string `yaml:"test_size_filters,omitempty"`
	TestTagFilters        string `yaml:"test_tag_filters,omitempty"`
	InstrumentationFilter string `yaml:"instrumentation_filter,omitempty"`
}
