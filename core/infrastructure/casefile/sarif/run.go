package sarif

type run struct {
	Tool    tool     `json:"tool"`
	Results []result `json:"results"`
}
