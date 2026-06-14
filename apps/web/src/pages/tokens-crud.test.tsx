import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { renderApp } from "@/test/render";
import { server } from "@/test/msw-server";
import { TokensPage } from "./tokens";

describe("Tokens — CRUD", () => {
  it("should load and display tokens from API", async () => {
    renderApp(<TokensPage />);
    expect(await screen.findByText("ci-token")).toBeInTheDocument();
    expect(screen.getByText("read-token")).toBeInTheDocument();
    expect(screen.getByText("gav_abc")).toBeInTheDocument();
    expect(screen.getByText("ingest")).toBeInTheDocument();
  });

  it("should open create form on button click", async () => {
    const user = userEvent.setup();
    renderApp(<TokensPage />);
    await screen.findByText("ci-token");

    await user.click(screen.getByRole("button", { name: /create token/i }));
    expect(screen.getByText("Create API token")).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
  });

  it("should show revealed token after creation (one time)", async () => {
    const user = userEvent.setup();
    renderApp(<TokensPage />);
    await screen.findByText("ci-token");

    await user.click(screen.getByRole("button", { name: /create token/i }));
    await user.type(screen.getByLabelText("Name"), "new-token");
    await user.click(screen.getByLabelText("ingest"));
    await user.click(screen.getByRole("button", { name: "Create" }));

    expect(await screen.findByText(/gav_test_full_token_value_here/)).toBeInTheDocument();
    expect(screen.getByText(/won't be shown again/)).toBeInTheDocument();
  });

  it("should hide reveal after closing", async () => {
    const user = userEvent.setup();
    renderApp(<TokensPage />);
    await screen.findByText("ci-token");

    await user.click(screen.getByRole("button", { name: /create token/i }));
    await user.type(screen.getByLabelText("Name"), "new-token");
    await user.click(screen.getByLabelText("ingest"));
    await user.click(screen.getByRole("button", { name: "Create" }));

    await screen.findByText(/gav_test_full_token_value_here/);
    await user.click(screen.getByRole("button", { name: /dismiss/i }));

    expect(screen.queryByText(/gav_test_full_token_value_here/)).not.toBeInTheDocument();
  });

  it("should delete token after confirmation", async () => {
    const user = userEvent.setup();
    renderApp(<TokensPage />);
    await screen.findByText("ci-token");

    const deleteButtons = screen.getAllByRole("button", { name: "" });
    await user.click(deleteButtons[0]);

    expect(screen.getByRole("button", { name: "Confirm" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Confirm" }));
  });

  it("should show error on API failure", async () => {
    server.use(
      http.post("/api/v1/me/tokens", () =>
        HttpResponse.json({ error: "Token name already exists" }, { status: 409 }),
      ),
    );

    const user = userEvent.setup();
    renderApp(<TokensPage />);
    await screen.findByText("ci-token");

    await user.click(screen.getByRole("button", { name: /create token/i }));
    await user.type(screen.getByLabelText("Name"), "ci-token");
    await user.click(screen.getByLabelText("ingest"));
    await user.click(screen.getByRole("button", { name: "Create" }));

    expect(await screen.findByText("Token name already exists")).toBeInTheDocument();
  });

  it("should cancel delete confirmation and restore delete icon", async () => {
    const user = userEvent.setup();
    renderApp(<TokensPage />);
    await screen.findByText("ci-token");

    const deleteButtons = screen.getAllByRole("button", { name: "" });
    await user.click(deleteButtons[0]);
    expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(screen.queryByRole("button", { name: "Confirm" })).not.toBeInTheDocument();
  });

  it("should close create form on cancel", async () => {
    const user = userEvent.setup();
    renderApp(<TokensPage />);
    await screen.findByText("ci-token");

    await user.click(screen.getByRole("button", { name: /create token/i }));
    expect(screen.getByText("Create API token")).toBeInTheDocument();

    const cancelButtons = screen.getAllByRole("button", { name: "Cancel" });
    await user.click(cancelButtons[cancelButtons.length - 1]);
    expect(screen.queryByText("Create API token")).not.toBeInTheDocument();
  });
});
