import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { api, postJson } from "../api/client";
import type { User } from "../types";

type AuthContextValue = {
  user: User | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  changePassword: (oldPassword: string, newPassword: string) => Promise<void>;
  refresh: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = async () => {
    try {
      const response = await api<{ user: User }>("/auth/me");
      setUser(response.user);
    } catch {
      setUser(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void refresh();
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      loading,
      login: async (username, password) => {
        const response = await postJson<{ user: User }>("/auth/login", { username, password });
        setUser(response.user);
      },
      logout: async () => {
        await postJson("/auth/logout", {});
        setUser(null);
      },
      changePassword: async (oldPassword, newPassword) => {
        const response = await postJson<{ user: User }>("/auth/change-password", {
          old_password: oldPassword,
          new_password: newPassword
        });
        setUser(response.user);
      },
      refresh
    }),
    [user, loading]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used inside AuthProvider");
  return context;
}
