import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EmptyState } from "./empty-state";

describe("EmptyState", () => {
  it("should render title and description", () => {
    render(
      <EmptyState title="No projects" description="Submit an analysis to get started." />,
    );
    expect(screen.getByText("No projects")).toBeInTheDocument();
    expect(screen.getByText("Submit an analysis to get started.")).toBeInTheDocument();
  });

  it("should render icon when provided", () => {
    const Icon = () => <svg data-testid="icon" />;
    render(<EmptyState icon={Icon} title="Empty" />);
    expect(screen.getByTestId("icon")).toBeInTheDocument();
  });

  it("should render action button when provided", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    render(
      <EmptyState
        title="No tokens"
        action={{ label: "Create token", onClick }}
      />,
    );

    const btn = screen.getByRole("button", { name: "Create token" });
    expect(btn).toBeInTheDocument();
    await user.click(btn);
    expect(onClick).toHaveBeenCalled();
  });

  it("should render without description or action", () => {
    render(<EmptyState title="Nothing here" />);
    expect(screen.getByText("Nothing here")).toBeInTheDocument();
  });
});
