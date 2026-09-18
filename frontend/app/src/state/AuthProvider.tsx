import {
  createContext,
  useContext,
  useState,
  type ReactNode,
} from "react";

const SESSION_KEY = "agentgate.admin-token";

interface AuthContextValue {
  isAuthenticated: boolean;
  login: (token: string) => { ok: boolean; error?: string };
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

/**
 * Minimal real gate matching how the backend actually authenticates today:
 * a single static admin bearer token (AGENTGATE_ADMIN_TOKEN), not per-user
 * login. There is no live backend connected yet, so there is nothing to
 * call — this checks the entered token against VITE_ADMIN_TOKEN when that
 * env var is set (matching what a real deployment would require), and
 * falls back to accepting any non-empty token for local/demo use when it
 * isn't set. Session state is a plain sessionStorage flag: it clears when
 * the tab closes, and is not a substitute for real server-side auth.
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(
    () => sessionStorage.getItem(SESSION_KEY) === "true"
  );

  const login = (token: string): { ok: boolean; error?: string } => {
    const trimmed = token.trim();
    if (trimmed.length === 0) {
      return { ok: false, error: "Enter the admin token." };
    }

    const expected = import.meta.env.VITE_ADMIN_TOKEN as string | undefined;
    if (expected && trimmed !== expected) {
      return { ok: false, error: "That token doesn't match." };
    }

    sessionStorage.setItem(SESSION_KEY, "true");
    setIsAuthenticated(true);
    return { ok: true };
  };

  const logout = () => {
    sessionStorage.removeItem(SESSION_KEY);
    setIsAuthenticated(false);
  };

  return (
    <AuthContext.Provider value={{ isAuthenticated, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth() must be used inside <AuthProvider>");
  }
  return ctx;
}
