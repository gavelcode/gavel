import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";

import { renderApp } from "@/test/render";
import { server } from "@/test/msw-server";

import { AdminUsersPage } from "./admin-users";

describe("AdminUsersPage", () => {
  it("creates a user via the API on submit and shows the success banner", async () => {
    let receivedBody: Record<string, unknown> | null = null;
    server.use(
      http.post("/api/v1/admin/users", async ({ request }) => {
        receivedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json({
          id: 42,
          email: "new@example.com",
          role: "maintainer",
        });
      }),
    );

    const user = userEvent.setup();
    renderApp(<AdminUsersPage />);

    await user.type(screen.getByLabelText(/email/i), "new@example.com");
    await user.type(screen.getByLabelText(/display name/i), "New User");
    await user.type(screen.getByLabelText(/password/i), "supersecret");
    await user.selectOptions(screen.getByLabelText(/role/i), "maintainer");
    await user.click(screen.getByRole("button", { name: /create user/i }));

    expect(
      await screen.findByText(/User new@example.com created with role maintainer/),
    ).toBeInTheDocument();
    expect(receivedBody).toEqual({
      email: "new@example.com",
      display_name: "New User",
      password: "supersecret",
      role: "maintainer",
    });
  });

  it("redirects non-admins away instead of rendering the form", () => {
    renderApp(<AdminUsersPage />, {
      auth: {
        user: {
          id: 2,
          email: "viewer@local",
          displayName: "Viewer",
          role: "viewer",
          mustChangePassword: false,
        },
      },
    });

    expect(screen.queryByRole("heading", { name: /create user/i })).not.toBeInTheDocument();
    expect(screen.queryByLabelText(/email/i)).not.toBeInTheDocument();
  });

  it("blocks submission and shows a local error when the password is too short", async () => {
    let apiCalled = false;
    server.use(
      http.post("/api/v1/admin/users", () => {
        apiCalled = true;
        return HttpResponse.json({ id: 1, email: "x", role: "viewer" });
      }),
    );

    const user = userEvent.setup();
    renderApp(<AdminUsersPage />);

    await user.type(screen.getByLabelText(/email/i), "short@example.com");
    await user.type(screen.getByLabelText(/display name/i), "Short");
    await user.type(screen.getByLabelText(/password/i), "1234567");
    await user.click(screen.getByRole("button", { name: /create user/i }));

    expect(
      await screen.findByText(/Password must be at least 8 characters/),
    ).toBeInTheDocument();
    expect(apiCalled).toBe(false);
  });

  it("surfaces the backend error detail when the API rejects the request", async () => {
    server.use(
      http.post("/api/v1/admin/users", () =>
        HttpResponse.json(
          { detail: "email already in use" },
          { status: 422 },
        ),
      ),
    );

    const user = userEvent.setup();
    renderApp(<AdminUsersPage />);

    await user.type(screen.getByLabelText(/email/i), "dup@example.com");
    await user.type(screen.getByLabelText(/display name/i), "Dup");
    await user.type(screen.getByLabelText(/password/i), "longenough");
    await user.click(screen.getByRole("button", { name: /create user/i }));

    expect(
      await screen.findByText(/email already in use/),
    ).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.queryByText(/User .* created with role/)).not.toBeInTheDocument();
    });
  });
});
