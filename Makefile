OAPI_CODEGEN_VERSION := v2.7.0
OAPI_CODEGEN := go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION)

V1_DIR := core/userinterface/api/v1
V1_GEN_DIR := $(V1_DIR)/gen
V1_SPEC := openapi/v1/openapi.yaml
V1_BUNDLE := /tmp/gavel-openapi-bundled.yaml

.PHONY: openapi-bundle openapi-gen openapi-gen-ts openapi-check clispec-gen clispec-check e2e release

# openapi-bundle uses @redocly/cli to inline the split spec under openapi/v1/
# into a single self-contained YAML. Both code generators (Go via oapi-codegen,
# TypeScript via openapi-typescript) consume this bundled artifact.
openapi-bundle:
	cd apps/web && npx redocly bundle ../../$(V1_SPEC) -o $(V1_BUNDLE)

openapi-gen: openapi-bundle openapi-gen-ts
	cd $(V1_GEN_DIR) && $(OAPI_CODEGEN) -config oapi-codegen.yaml $(V1_BUNDLE)

openapi-gen-ts: openapi-bundle
	cd apps/web && pnpm run openapi-gen

openapi-check: openapi-gen
	@git diff --exit-code -- $(V1_DIR) apps/web/src/shared/api/v1.gen.ts \
		|| (echo "openapi drift detected: run 'make openapi-gen' and commit"; exit 1)

clispec-gen:
	cd tools && go run ./clispec-gen ../clispec/v1/clispec.yaml

clispec-check: clispec-gen
	@git diff --exit-code -- core/userinterface/cli/*/flags.gen.go \
		|| (echo "clispec drift detected: run 'make clispec-gen' and commit"; exit 1)
	bazel test //apps/cli/test/integration/clispec/...

e2e:
	bazel build //apps/server/cmd/gavel-server
	cd apps/web && npx playwright test

# release tags and publishes a CLI release. It only validates and tags; the
# release workflow does the build/publish. Usage: make release VERSION=X.Y.Z
release:
	@test -n "$(VERSION)" || { echo "usage: make release VERSION=X.Y.Z"; exit 1; }
	@bash hack/release.sh $(VERSION)
