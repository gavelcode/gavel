import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderApp } from "@/test/render";
import "@/test/msw-server";
import { CreateTokenForm } from "./create-token-form";

describe("CreateTokenForm", () => {
  it("should show error when submitting without selecting any scope", async () => {
    const user = userEvent.setup();
    const onCreated = vi.fn();
    const onCancel = vi.fn();
    renderApp(<CreateTokenForm onCreated={onCreated} onCancel={onCancel} />);

    await user.type(screen.getByLabelText("Name"), "test-token");
    await user.click(screen.getByRole("button", { name: "Create" }));

    expect(screen.getByText("Select at least one scope")).toBeInTheDocument();
    expect(onCreated).not.toHaveBeenCalled();
  });

  it("should update expiration days input when typing", async () => {
    const user = userEvent.setup();
    renderApp(<CreateTokenForm onCreated={vi.fn()} onCancel={vi.fn()} />);

    const expiresInput = screen.getByLabelText(/expires in/i);
    await user.type(expiresInput, "30");
    expect(expiresInput).toHaveValue(30);
  });
});
