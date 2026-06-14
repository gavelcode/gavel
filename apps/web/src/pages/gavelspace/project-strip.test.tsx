import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { ProjectStrip } from "./project-strip";
import type { ProjectRef } from "@/entities/gavelspace/model";

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return { ...actual, useNavigate: () => mockNavigate };
});

describe("ProjectStrip", () => {
  beforeEach(() => {
    mockNavigate.mockClear();
  });

  it("renders nothing when projects array is empty", () => {
    const { container } = render(
      <MemoryRouter>
        <ProjectStrip projects={[]} />
      </MemoryRouter>,
    );

    expect(container.innerHTML).toBe("");
  });

  it("renders project cards when projects are provided", () => {
    const projects: ProjectRef[] = [
      { id: "1", key: "core", name: "Core", latestVerdict: "pass" },
    ];

    render(
      <MemoryRouter>
        <ProjectStrip projects={projects} />
      </MemoryRouter>,
    );

    expect(screen.getByText("Core")).toBeInTheDocument();
    expect(screen.getByText("core")).toBeInTheDocument();
  });

  it("navigates to project page on card click", async () => {
    const projects: ProjectRef[] = [
      { id: "1", key: "core", name: "Core", latestVerdict: "pass" },
    ];
    const user = userEvent.setup();

    render(
      <MemoryRouter>
        <ProjectStrip projects={projects} />
      </MemoryRouter>,
    );

    await user.click(screen.getByTestId("project-strip-card"));

    expect(mockNavigate).toHaveBeenCalledWith("/projects/core");
  });
});
