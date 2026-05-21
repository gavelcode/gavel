package collector_test

import (
	"context"

	"github.com/usegavel/gavel/core/infrastructure/platform/bazel/runner"
	"github.com/usegavel/gavel/core/application/casefile/evidencedto"
	ingestfind "github.com/usegavel/gavel/core/application/casefile/ingestfindings"
)

type fakeAnalysisRunner struct {
	sarifFiles   map[string][][]byte
	coverageData []byte
	buildWarning error
	err          error
}

func (f *fakeAnalysisRunner) RunAnalysis(_ context.Context, _ runner.AnalysisConfig) (*runner.AnalysisResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &runner.AnalysisResult{
		SARIFFiles:   f.sarifFiles,
		CoverageData: f.coverageData,
		BuildWarning: f.buildWarning,
	}, nil
}

type fakeFindingsParser struct {
	returnEmpty bool
	err         error
}

func (f *fakeFindingsParser) Execute(_ context.Context, _ ingestfind.Command) (ingestfind.Result, error) {
	if f.err != nil {
		return ingestfind.Result{}, f.err
	}
	return ingestfind.Result{Evidence: evidencedto.Evidence{Subtype: "code_quality"}}, nil
}
