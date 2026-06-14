import { screen, render } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw-server";
import { AuthProvider } from "./auth-provider";
import { useAuth } from "./use-auth";

function TestConsumer() {
  const { user, loading, login, logout } = useAuth();
  if (loading) return <div>loading</div>;
  if (!user) return (
    <div>
      <span>not authenticated</span>
      <button onClick={() => login("admin@local", "admin123!")}>login</button>
    </div>
  );
  return (
    <div>
      <span>hello {user.displayName}</span>
      <button onClick={() => logout()}>logout</button>
    </div>
  );
}

function renderWithProviders(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>{ui}</MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("AuthProvider", () => {
  it("loads the current user on mount", async () => {
    renderWithProviders(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    );
    expect(screen.getByText("loading")).toBeInTheDocument();
    expect(await screen.findByText("hello Admin")).toBeInTheDocument();
  });

  it("sets user to null when /me fails", async () => {
    server.use(
      http.get("/api/v1/me", () => HttpResponse.json({ error: "Unauthorized" }, { status: 401 })),
    );
    renderWithProviders(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    );
    expect(await screen.findByText("not authenticated")).toBeInTheDocument();
  });

  it("login sets user on success", async () => {
    server.use(
      http.get("/api/v1/me", () => HttpResponse.json({ error: "Unauthorized" }, { status: 401 })),
    );
    renderWithProviders(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    );
    await screen.findByText("not authenticated");
    await userEvent.click(screen.getByText("login"));
    expect(await screen.findByText("hello Admin")).toBeInTheDocument();
  });

  it("logout clears user", async () => {
    renderWithProviders(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    );
    await screen.findByText("hello Admin");
    await userEvent.click(screen.getByText("logout"));
    expect(await screen.findByText("not authenticated")).toBeInTheDocument();
  });
});
