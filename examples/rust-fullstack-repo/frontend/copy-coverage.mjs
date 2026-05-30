import { existsSync, copyFileSync } from "fs";

export function teardown() {
  const outFile = process.env.COVERAGE_OUTPUT_FILE;
  const coverageDir = process.env.COVERAGE_DIR;
  if (outFile && coverageDir) {
    const lcov = coverageDir + "/lcov.info";
    if (existsSync(lcov)) {
      copyFileSync(lcov, outFile);
    }
  }
}
