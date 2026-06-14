import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { renderApp } from "@/test/render";
import { server } from "@/test/msw-server";
import { LoginPage } from "./login";
import { Routes, Route } from "react-router-dom";

function renderLogin(authenticated = false) {
  const auth = authenticated
    ? {}
    : {
        user: null,
        loading: false,
        login: async (email: string, password: string) => {
          const { request } = await import("@/shared/api/client");
          await request("/api/v1/sessions", {
            method: "POST",
            body: JSON.stringify({ email, password }),
          });
        },
      };

  return renderApp(
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<div>Projects</div>} />
    </Routes>,
    { auth, route: "/login" },
  );
}

describe("Login", () => {
  it("should show login form with email and password fields", () => {
    renderLogin();
    expect(screen.getByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /sign in/i })).toBeInTheDocument();
  });

  it("should redirect to / on successful login", async () => {
    const user = userEvent.setup();
    renderLogin();

    await user.clear(screen.getByLabelText("Email"));
    await user.type(screen.getByLabelText("Email"), "admin@local");
    await user.clear(screen.getByLabelText("Password"));
    await user.type(screen.getByLabelText("Password"), "admin123!");
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    expect(await screen.findByText("Projects")).toBeInTheDocument();
  });

  it("should show error on invalid credentials", async () => {
    const user = userEvent.setup();
    renderLogin();

    await user.clear(screen.getByLabelText("Email"));
    await user.type(screen.getByLabelText("Email"), "wrong@email.com");
    await user.clear(screen.getByLabelText("Password"));
    await user.type(screen.getByLabelText("Password"), "wrongpass");
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    expect(await screen.findByText("Invalid credentials")).toBeInTheDocument();
  });

  it("should show error on connection failure", async () => {
    server.use(
      http.post("/api/v1/sessions", () => HttpResponse.error()),
    );

    const user = userEvent.setup();
    renderLogin();

    await user.clear(screen.getByLabelText("Email"));
    await user.type(screen.getByLabelText("Email"), "admin@local");
    await user.clear(screen.getByLabelText("Password"));
    await user.type(screen.getByLabelText("Password"), "admin123!");
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    expect(await screen.findByText("Connection error")).toBeInTheDocument();
  });

  it("should redirect to / if already authenticated", () => {
    renderLogin(true);
    expect(screen.getByText("Projects")).toBeInTheDocument();
    expect(screen.queryByLabelText("Email")).not.toBeInTheDocument();
  });

  it("should disable button while submitting", async () => {
    server.use(
      http.post("/api/v1/sessions", () => new Promise(() => {})),
    );

    const user = userEvent.setup();
    renderLogin();

    await user.clear(screen.getByLabelText("Email"));
    await user.type(screen.getByLabelText("Email"), "admin@local");
    await user.clear(screen.getByLabelText("Password"));
    await user.type(screen.getByLabelText("Password"), "admin123!");
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    expect(screen.getByRole("button", { name: /sign in/i })).toBeDisabled();
  });
});
