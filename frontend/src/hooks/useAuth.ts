import { useCallback, useSyncExternalStore } from "react";
import type { TokenResponse } from "../api/types";
import * as authApi from "../api/auth";

const TOKEN_KEY = "tokens";
const USER_KEY = "current_user";

function getStoredTokens(): TokenResponse | null {
  try {
    const raw = localStorage.getItem(TOKEN_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

function getStoredUserId(): number | null {
  try {
    const raw = localStorage.getItem(USER_KEY);
    return raw ? (JSON.parse(raw) as number) : null;
  } catch {
    return null;
  }
}

let listeners: (() => void)[] = [];

function subscribe(cb: () => void) {
  listeners.push(cb);
  return () => {
    listeners = listeners.filter((l) => l !== cb);
  };
}

function emit() {
  listeners.forEach((l) => l());
}

export function useAuth() {
  const tokens = useSyncExternalStore(subscribe, getStoredTokens);
  const userId = useSyncExternalStore(subscribe, getStoredUserId);
  const isAuthenticated = tokens !== null;

  const saveSession = useCallback(
    (tokenResp: TokenResponse, uid: number) => {
      localStorage.setItem(TOKEN_KEY, JSON.stringify(tokenResp));
      localStorage.setItem(USER_KEY, JSON.stringify(uid));
      emit();
    },
    [],
  );

  const login = useCallback(
    async (username: string, password: string) => {
      const res = await authApi.login({ username, password });
      if (!res.data) throw new Error("login failed");
      // 登录后先存 tokens，再拿用户信息
      localStorage.setItem(TOKEN_KEY, JSON.stringify(res.data));
      emit();
      return res.data;
    },
    [],
  );

  const setUserId = useCallback((id: number) => {
    localStorage.setItem(USER_KEY, JSON.stringify(id));
    emit();
  }, []);

  const register = useCallback(
    async (username: string, password: string, email?: string, nickname?: string) => {
      const res = await authApi.register({ username, password, email, nickname });
      if (!res.data) throw new Error("register failed");
      return res.data;
    },
    [],
  );

  const logout = useCallback(async () => {
    const t = getStoredTokens();
    if (t) {
      try {
        await authApi.logout(t.access_token, t.refresh_token);
      } catch {
        // 即使后端调用失败也清除本地状态
      }
    }
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    emit();
  }, []);

  return {
    tokens,
    userId,
    isAuthenticated,
    login,
    register,
    logout,
    saveSession,
    setUserId,
  };
}
