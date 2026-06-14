import { render, type RenderOptions } from "@testing-library/react";
import { MemoryRouter, type MemoryRouterProps } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext, type AuthState } from "@/entities/user/use-auth";
import { ThemeProvider } from "@/shared/lib/theme";
import { DensityProvider } from "@/shared/lib/density";
import type { User } from "@/entities/user/model";

const defaultUser: User = {
  id: 1,
  email: "admin@local",
  displayName: "Admin",
  role: "admin",
  mustChangePassword: false,
};

const defaultAuth: AuthState = {
  user: defaultUser,
  loading: false,
  login: async () => {},
  logout: async () => {},
};

interface AppRenderOptions extends Omit<RenderOptions, "wrapper"> {
  auth?: Partial<AuthState>;
  route?: string;
  routerProps?: MemoryRouterProps;
}

export function renderApp(ui: React.ReactElement, options: AppRenderOptions = {}) {
  const { auth = {}, route = "/", routerProps, ...renderOptions } = options;

  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  const authState: AuthState = { ...defaultAuth, ...auth };

  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <AuthContext.Provider value={authState}>
          <ThemeProvider>
            <DensityProvider>
              <MemoryRouter initialEntries={[route]} {...routerProps}>
                {children}
              </MemoryRouter>
            </DensityProvider>
          </ThemeProvider>
        </AuthContext.Provider>
      </QueryClientProvider>
    );
  }

  return { ...render(ui, { wrapper: Wrapper, ...renderOptions }), queryClient };
}
