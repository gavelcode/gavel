package ingestncc

import "github.com/usegavel/gavel/core/application/casefile/evidencedto"

type Result struct {
	Evidence evidencedto.Evidence
	Percent  float64
}
