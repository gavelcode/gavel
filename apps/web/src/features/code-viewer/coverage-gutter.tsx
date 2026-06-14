export type CoverageState = "covered" | "uncovered" | "none";

export interface CoverageGutterProps {
  state: CoverageState;
}

const stateClassName: Record<CoverageState, string> = {
  covered: "bg-success/60",
  uncovered: "bg-danger/60",
  none: "",
};

const stateAriaLabel: Record<CoverageState, string> = {
  covered: "Coverage: covered",
  uncovered: "Coverage: uncovered",
  none: "Coverage: no data",
};

export function CoverageGutter({ state }: CoverageGutterProps) {
  return (
    <span
      data-testid="coverage-gutter"
      data-coverage={state}
      aria-label={stateAriaLabel[state]}
      className={`block w-1 self-stretch ${stateClassName[state]}`}
    />
  );
}
