package ingestncc

type PerLineParser interface {
	ParsePerLine(data []byte) (map[string]map[int]int, error)
}
