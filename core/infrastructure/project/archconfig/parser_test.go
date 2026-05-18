package archconfig_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/shared/failure"
	"github.com/usegavel/gavel/core/infrastructure/project/archconfig"
)

func TestParseShouldProducePolicyFromV2Config(t *testing.T) {
	data := []byte(`
layers:
  domain: ["internal/domain/..."]
  application: ["internal/application/..."]
  infrastructure: ["internal/infrastructure/..."]
rules:
  - name: domain-imports-nothing
    source: domain
    deny: [application, infrastructure]
detect_cycles: true
`)

	policy, err := archconfig.Parse(data)
	require.NoError(t, err)

	assert.Len(t, policy.Layers(), 3)
	assert.Len(t, policy.DenyRules(), 1)
	assert.True(t, policy.DetectCycles())

	rule := policy.DenyRules()[0]
	assert.Equal(t, "domain-imports-nothing", rule.Name())
	assert.Equal(t, "domain", rule.Source())
	assert.Equal(t, []string{"application", "infrastructure"}, rule.Deny())
}

func TestParseShouldProducePolicyFromV1Config(t *testing.T) {
	data := []byte(`
version: 1
module: github.com/example/repo
layers:
  domain: ["internal/domain/..."]
  application: ["internal/application/..."]
rules:
  - name: domain-purity
    source: domain
    deny: [application]
generic:
  no_circular_deps: true
`)

	policy, err := archconfig.Parse(data)
	require.NoError(t, err)

	assert.Len(t, policy.Layers(), 2)
	assert.Len(t, policy.DenyRules(), 1)
	assert.True(t, policy.DetectCycles())
}

func TestParseShouldAcceptPolicyWithoutRules(t *testing.T) {
	data := []byte(`
layers:
  domain: ["internal/domain/..."]
  application: ["internal/application/..."]
`)

	policy, err := archconfig.Parse(data)
	require.NoError(t, err)

	assert.Len(t, policy.Layers(), 2)
	assert.Empty(t, policy.DenyRules())
	assert.False(t, policy.DetectCycles())
}

func TestParseShouldRejectEmptyLayers(t *testing.T) {
	data := []byte(`
layers: {}
`)

	_, err := archconfig.Parse(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, archconfig.ErrParseConfig,
		"empty layers must classify as ErrParseConfig")
}

func TestParseShouldRejectRuleReferencingUnknownLayer(t *testing.T) {
	data := []byte(`
layers:
  domain: ["internal/domain/..."]
rules:
  - name: bad-rule
    source: nonexistent
    deny: [domain]
`)

	_, err := archconfig.Parse(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, archconfig.ErrParseConfig,
		"unknown layer reference must classify as ErrParseConfig")
}

func TestParseShouldRejectInvalidYAML(t *testing.T) {
	data := []byte(`{{{invalid`)

	_, err := archconfig.Parse(data)
	assert.Error(t, err)
}

func TestParseShouldSupportMultiplePatterns(t *testing.T) {
	data := []byte(`
layers:
  domain:
    - "internal/domain/..."
    - "pkg/domain/..."
  application: ["internal/application/..."]
`)

	policy, err := archconfig.Parse(data)
	require.NoError(t, err)

	for _, l := range policy.Layers() {
		if l.Name() == "domain" {
			assert.Len(t, l.Patterns(), 2)
			return
		}
	}
	t.Fatal("domain layer not found")
}

func TestParseShouldSupportMultipleRules(t *testing.T) {
	data := []byte(`
layers:
  domain: ["internal/domain/..."]
  application: ["internal/application/..."]
  infrastructure: ["internal/infrastructure/..."]
rules:
  - name: domain-purity
    source: domain
    deny: [application, infrastructure]
  - name: app-no-infra
    source: application
    deny: [infrastructure]
`)

	policy, err := archconfig.Parse(data)
	require.NoError(t, err)

	assert.Len(t, policy.DenyRules(), 2)
}

func TestParseInvalidYAMLReturnsParseConfigError(t *testing.T) {
	_, err := archconfig.Parse([]byte(`{invalid yaml`))
	require.Error(t, err)
	assert.ErrorIs(t, err, archconfig.ErrParseConfig)
	assert.Equal(t, failure.Validation, failure.Of(err))
}

func TestParseFileShouldParsePolicyFromValidFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/architecture.yml"
	content := []byte(`
layers:
  domain: ["internal/domain/..."]
  application: ["internal/application/..."]
rules:
  - name: domain-purity
    source: domain
    deny: [application]
`)
	require.NoError(t, os.WriteFile(path, content, 0o644))

	policy, err := archconfig.ParseFile(path)
	require.NoError(t, err)

	assert.Len(t, policy.Layers(), 2)
	assert.Len(t, policy.DenyRules(), 1)
}

func TestParseFileShouldReturnReadConfigErrorForNonExistentFile(t *testing.T) {
	_, err := archconfig.ParseFile("/nonexistent/path/architecture.yml")
	require.Error(t, err)
	assert.ErrorIs(t, err, archconfig.ErrReadConfig)
}

func TestParseShouldRejectLayerWithEmptyPatterns(t *testing.T) {
	data := []byte(`
layers:
  domain: []
  application: ["internal/application/..."]
rules:
  - name: domain-purity
    source: domain
    deny: [application]
`)

	_, err := archconfig.Parse(data)
	require.Error(t, err)
	assert.ErrorIs(t, err, archconfig.ErrParseConfig)
}

func TestParseShouldRejectLayerWithEmptyName(t *testing.T) {
	data := []byte(`
layers:
  "": ["internal/domain/..."]
  application: ["internal/application/..."]
`)

	_, err := archconfig.Parse(data)
	require.Error(t, err)
	assert.ErrorIs(t, err, archconfig.ErrParseConfig)
}

func TestParseShouldRejectRuleWithEmptyName(t *testing.T) {
	data := []byte(`
layers:
  domain: ["internal/domain/..."]
  application: ["internal/application/..."]
rules:
  - name: ""
    source: domain
    deny: [application]
`)

	_, err := archconfig.Parse(data)
	require.Error(t, err)
	assert.ErrorIs(t, err, archconfig.ErrParseConfig)
}

func TestParseShouldRejectRuleWithEmptySource(t *testing.T) {
	data := []byte(`
layers:
  domain: ["internal/domain/..."]
  application: ["internal/application/..."]
rules:
  - name: domain-purity
    source: ""
    deny: [application]
`)

	_, err := archconfig.Parse(data)
	require.Error(t, err)
	assert.ErrorIs(t, err, archconfig.ErrParseConfig)
}

func TestParseShouldRejectRuleWithEmptyDenyList(t *testing.T) {
	data := []byte(`
layers:
  domain: ["internal/domain/..."]
  application: ["internal/application/..."]
rules:
  - name: domain-purity
    source: domain
    deny: []
`)

	_, err := archconfig.Parse(data)
	require.Error(t, err)
	assert.ErrorIs(t, err, archconfig.ErrParseConfig)
}
