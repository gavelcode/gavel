import { useState, useEffect, useCallback, type ReactNode } from "react";
import { ApiErrorResponse } from "@/shared/api/client";
import type { User } from "./model";
import { AuthContext } from "./use-auth";
import * as userApi from "./api";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    userApi
      .me()
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const loggedIn = await userApi.login(email, password);
    setUser(loggedIn);
  }, []);

  const logout = useCallback(async () => {
    try {
      await userApi.logout();
    } catch (e) {
      if (!(e instanceof ApiErrorResponse && e.status === 401)) throw e;
    }
    setUser(null);
  }, []);

  return (
    <AuthContext value={{ user, loading, login, logout }}>
      {children}
    </AuthContext>
  );
}
