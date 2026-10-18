import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { api, applySession, hasStoredSession, refreshSession, setSessionListener } from "./api";
import type { AuthResponse, User } from "./types";

interface AuthState {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string, role: "buyer" | "seller") => Promise<void>;
  logout: () => Promise<void>;
  reload: () => Promise<boolean>;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(hasStoredSession());

  useEffect(() => {
    setSessionListener((auth) => setUser(auth?.user ?? null));
    if (hasStoredSession()) refreshSession().finally(() => setLoading(false));
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    applySession(await api<AuthResponse>("/auth/login", { method: "POST", json: { email, password } }));
  }, []);

  const register = useCallback(async (username: string, email: string, password: string, role: "buyer" | "seller") => {
    applySession(await api<AuthResponse>("/auth/register", { method: "POST", json: { username, email, password, role } }));
  }, []);

  const logout = useCallback(async () => {
    const token = localStorage.getItem("mp.refresh");
    applySession(null);
    if (token) await api("/auth/logout", { method: "POST", json: { refresh_token: token } }).catch(() => undefined);
  }, []);

  const value = useMemo(
    () => ({ user, loading, login, register, logout, reload: refreshSession }),
    [user, loading, login, register, logout],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside AuthProvider");
  return ctx;
}
