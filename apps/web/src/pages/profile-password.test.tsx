import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { renderApp } from "@/test/render";
import { server } from "@/test/msw-server";
import { ProfilePage } from "./profile";

describe("Profile — Change Password", () => {
  it("should show current user info (email, name, role)", () => {
    renderApp(<ProfilePage />);
    expect(screen.getByText("admin@local")).toBeInTheDocument();
    expect(screen.getByText("Admin")).toBeInTheDocument();
    expect(screen.getByText("admin")).toBeInTheDocument();
  });

  it("should validate passwords match before submit", async () => {
    const user = userEvent.setup();
    renderApp(<ProfilePage />);

    await user.type(screen.getByLabelText("Current password"), "oldpass1");
    await user.type(screen.getByLabelText("New password"), "newpass123");
    await user.type(screen.getByLabelText("Confirm new password"), "different1");
    await user.click(screen.getByRole("button", { name: /change password/i }));

    expect(screen.getByText("Passwords do not match")).toBeInTheDocument();
  });

  it("should validate minimum password length", async () => {
    const user = userEvent.setup();
    renderApp(<ProfilePage />);

    await user.type(screen.getByLabelText("Current password"), "oldpass1");
    await user.type(screen.getByLabelText("New password"), "short");
    await user.type(screen.getByLabelText("Confirm new password"), "short");
    await user.click(screen.getByRole("button", { name: /change password/i }));

    expect(screen.getByText("New password must be at least 8 characters")).toBeInTheDocument();
  });

  it("should show success message after change", async () => {
    const user = userEvent.setup();
    renderApp(<ProfilePage />);

    await user.type(screen.getByLabelText("Current password"), "oldpass12");
    await user.type(screen.getByLabelText("New password"), "newpass123");
    await user.type(screen.getByLabelText("Confirm new password"), "newpass123");
    await user.click(screen.getByRole("button", { name: /change password/i }));

    expect(await screen.findByText("Password changed successfully")).toBeInTheDocument();
  });

  it("should show error on API failure", async () => {
    server.use(
      http.post("/api/v1/me/password", () =>
        HttpResponse.json({ error: "Current password is incorrect" }, { status: 400 }),
      ),
    );

    const user = userEvent.setup();
    renderApp(<ProfilePage />);

    await user.type(screen.getByLabelText("Current password"), "wrongold1");
    await user.type(screen.getByLabelText("New password"), "newpass123");
    await user.type(screen.getByLabelText("Confirm new password"), "newpass123");
    await user.click(screen.getByRole("button", { name: /change password/i }));

    expect(await screen.findByText("Current password is incorrect")).toBeInTheDocument();
  });

  it("should clear form after successful change", async () => {
    const user = userEvent.setup();
    renderApp(<ProfilePage />);

    await user.type(screen.getByLabelText("Current password"), "oldpass12");
    await user.type(screen.getByLabelText("New password"), "newpass123");
    await user.type(screen.getByLabelText("Confirm new password"), "newpass123");
    await user.click(screen.getByRole("button", { name: /change password/i }));

    await screen.findByText("Password changed successfully");
    expect(screen.getByLabelText("Current password")).toHaveValue("");
    expect(screen.getByLabelText("New password")).toHaveValue("");
    expect(screen.getByLabelText("Confirm new password")).toHaveValue("");
  });
});
