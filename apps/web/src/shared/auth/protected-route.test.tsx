import { screen } from "@testing-library/react";
import "@/test/msw-server";
import { renderApp } from "@/test/render";
import { ProtectedRoute } from "./protected-route";
import { Routes, Route } from "react-router-dom";

describe("ProtectedRoute", () => {
  it("should render children when authenticated", () => {
    renderApp(
      <Routes>
        <Route element={<ProtectedRoute />}>
          <Route path="/" element={<div>Protected content</div>} />
        </Route>
      </Routes>,
    );
    expect(screen.getByText("Protected content")).toBeInTheDocument();
  });

  it("should redirect to /login when not authenticated", () => {
    renderApp(
      <Routes>
        <Route path="/login" element={<div>Login page</div>} />
        <Route element={<ProtectedRoute />}>
          <Route path="/" element={<div>Protected content</div>} />
        </Route>
      </Routes>,
      { auth: { user: null, loading: false } },
    );
    expect(screen.getByText("Login page")).toBeInTheDocument();
    expect(screen.queryByText("Protected content")).not.toBeInTheDocument();
  });

  it("should show spinner while auth is loading", () => {
    renderApp(
      <Routes>
        <Route element={<ProtectedRoute />}>
          <Route path="/" element={<div>Protected content</div>} />
        </Route>
      </Routes>,
      { auth: { user: null, loading: true } },
    );
    expect(screen.getByText("Loading...")).toBeInTheDocument();
    expect(screen.queryByText("Protected content")).not.toBeInTheDocument();
  });
});
