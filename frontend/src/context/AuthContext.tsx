import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";

const AUTH_KEY = "tasktracker_auth";

export type User = { username: string; user_id: number };

type AuthState = {
  token: string | null;
  user: User | null;
  ready: boolean;
};

type AuthContextValue = AuthState & {
  setAuth: (token: string, user: User) => void;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

function loadStored(): { token: string; user: User } | null {
  try {
    const raw = localStorage.getItem(AUTH_KEY);
    if (!raw) return null;
    const data = JSON.parse(raw) as { token: string; user: User };
    if (data.token && data.user?.username != null && data.user?.user_id != null) {
      return data;
    }
  } catch {
    // ignore
  }
  return null;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    token: null,
    user: null,
    ready: false,
  });

  useEffect(() => {
    const stored = loadStored();
    if (stored) {
      setState({ token: stored.token, user: stored.user, ready: true });
    } else {
      setState((s) => ({ ...s, ready: true }));
    }
  }, []);

  const setAuth = useCallback((token: string, user: User) => {
    const payload = { token, user };
    localStorage.setItem(AUTH_KEY, JSON.stringify(payload));
    setState({ token, user, ready: true });
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(AUTH_KEY);
    setState({ token: null, user: null, ready: true });
  }, []);

  const value: AuthContextValue = {
    ...state,
    setAuth,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (ctx == null) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
