package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/project/model"
)

func TestNewBridge(t *testing.T) {
	bridge, err := model.NewBridge("shellcheck", "//tools/lint:shellcheck.bzl%shellcheck_gavel", ".shellcheck.sarif", []string{"--some=flag"})

	require.NoError(t, err)
	assert.Equal(t, "shellcheck", bridge.Name())
	assert.Equal(t, "//tools/lint:shellcheck.bzl%shellcheck_gavel", bridge.Aspect())
	assert.Equal(t, ".shellcheck.sarif", bridge.SARIFSuffix())
	assert.Equal(t, []string{"--some=flag"}, bridge.BuildFlags())
}

func TestNewBridgeValidation(t *testing.T) {
	const aspect = "//tools/lint:a.bzl%a"
	tests := []struct {
		name    string
		bridge  string
		aspect  string
		suffix  string
		flags   []string
		message string
	}{
		{name: "shouldRejectMissingName", aspect: aspect, suffix: ".a.sarif", message: "name"},
		{name: "shouldRejectBlankName", bridge: "  ", aspect: aspect, suffix: ".a.sarif", message: "name"},
		{name: "shouldRejectNameWithWhitespace", bridge: "my tool", aspect: aspect, suffix: ".a.sarif", message: "name"},
		{name: "shouldRejectMissingAspect", bridge: "a", suffix: ".a.sarif", message: "aspect"},
		{name: "shouldRejectAspectWithoutPercent", bridge: "a", aspect: "//tools/lint:a.bzl", suffix: ".a.sarif", message: "%"},
		{name: "shouldRejectAspectWithEmptySymbol", bridge: "a", aspect: "//tools/lint:a.bzl%", suffix: ".a.sarif", message: "%"},
		{name: "shouldRejectMissingSuffix", bridge: "a", aspect: aspect, message: "sarif_suffix"},
		{name: "shouldRejectSuffixNotEndingInSarif", bridge: "a", aspect: aspect, suffix: ".a.json", message: ".sarif"},
		{name: "shouldRejectBareSarifSuffix", bridge: "a", aspect: aspect, suffix: ".sarif", message: ".sarif"},
		{name: "shouldRejectEmptyBuildFlag", bridge: "a", aspect: aspect, suffix: ".a.sarif", flags: []string{""}, message: "build_flags"},
		{name: "shouldRejectBuildFlagWithoutDash", bridge: "a", aspect: aspect, suffix: ".a.sarif", flags: []string{"//x"}, message: "build_flags"},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			_, err := model.NewBridge(tcase.bridge, tcase.aspect, tcase.suffix, tcase.flags)

			require.Error(t, err)
			assert.ErrorIs(t, err, model.ErrInvalidBridge)
			assert.Contains(t, err.Error(), tcase.message)
		})
	}
}

func TestNewBridgeCopiesBuildFlags(t *testing.T) {
	flags := []string{"--a"}
	bridge, err := model.NewBridge("a", "//t:a.bzl%a", ".a.sarif", flags)
	require.NoError(t, err)

	flags[0] = "--tampered"
	bridge.BuildFlags()[0] = "--tampered"

	assert.Equal(t, []string{"--a"}, bridge.BuildFlags())
}

func mustBridge(t *testing.T, name string) model.Bridge {
	t.Helper()
	bridge, err := model.NewBridge(name, "//tools/lint:"+name+".bzl%"+name, "."+name+".sarif", nil)
	require.NoError(t, err)
	return bridge
}

func TestProjectUpdateBridges(t *testing.T) {
	project, err := model.NewProject(testTenant, "payments", "payments", "//payments/...")
	require.NoError(t, err)
	project.ClearEvents()

	require.NoError(t, project.UpdateBridges([]model.Bridge{mustBridge(t, "shellcheck")}, testTime))

	require.Len(t, project.Bridges(), 1)
	assert.Equal(t, "shellcheck", project.Bridges()[0].Name())
	require.Len(t, project.Events(), 1)
	assert.Equal(t, model.EventNameBridgesUpdated, project.Events()[0].EventName())
}

func TestProjectUpdateBridgesRejectsDuplicateNames(t *testing.T) {
	project, err := model.NewProject(testTenant, "payments", "payments", "//payments/...")
	require.NoError(t, err)

	err = project.UpdateBridges([]model.Bridge{mustBridge(t, "shellcheck"), mustBridge(t, "shellcheck")}, testTime)

	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrInvalidBridge)
	assert.Contains(t, err.Error(), "duplicate")
	assert.Empty(t, project.Bridges())
}

func TestProjectUpdateBridgesRejectsClashWithCatalogToolOfTheProject(t *testing.T) {
	project, err := model.NewProject(testTenant, "payments", "payments", "//payments/...")
	require.NoError(t, err)
	project.UpdateToolSelection(map[string][]string{"go": {"golangci-lint"}}, testTime)

	err = project.UpdateBridges([]model.Bridge{mustBridge(t, "golangci-lint")}, testTime)

	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrInvalidBridge)
	assert.Contains(t, err.Error(), "golangci-lint")
}

func TestProjectBridgesReturnsDefensiveCopy(t *testing.T) {
	project, err := model.NewProject(testTenant, "payments", "payments", "//payments/...")
	require.NoError(t, err)
	require.NoError(t, project.UpdateBridges([]model.Bridge{mustBridge(t, "shellcheck")}, testTime))

	project.Bridges()[0] = mustBridge(t, "tampered")

	assert.Equal(t, "shellcheck", project.Bridges()[0].Name())
}
