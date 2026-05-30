import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    environment: "jsdom",
    coverage: {
      provider: "v8",
      reporter: ["lcov"],
      reportsDirectory: process.env.COVERAGE_DIR || "./coverage",
    },
    globalSetup: ["./copy-coverage.mjs"],
  },
  esbuild: {
    jsx: "automatic",
  },
});
