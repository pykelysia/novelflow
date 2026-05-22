import { useCallback, useSyncExternalStore } from "react";
import type { TokenResponse } from "../api/types";
import * as authApi from "../api/auth";

const TOKEN_KEY = "tokens";
const USER_KEY = "current_user";

// ---- 直接从 localStorage 读取（内部使用） ----

function readTokens(): TokenResponse | null {
  try {
    const raw = localStorage.getItem(TOKEN_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

function readUserId(): number | null {
  try {
    const raw = localStorage.getItem(USER_KEY);
    return raw ? (JSON.parse(raw) as number) : null;
  } catch {
    return null;
  }
}

// ---- 缓存快照，保证 useSyncExternalStore 的 getSnapshot 返回稳定引用 ----

let cachedTokens: TokenResponse | null = null;
let cachedUserId: number | null = null;

function syncCache() {
  cachedTokens = readTokens();
  cachedUserId = readUserId();
}

syncCache();

// ---- 订阅机制 ----

let listeners: (() => void)[] = [];

function subscribe(cb: () => void) {
  listeners.push(cb);
  return () => {
    listeners = listeners.filter((l) => l !== cb);
  };
}

function emit() {
  syncCache();
  listeners.forEach((l) => l());
}

// ---- Hook ----

export function useAuth() {
  const tokens = useSyncExternalStore(subscribe, () => cachedTokens);
  const userId = useSyncExternalStore(subscribe, () => cachedUserId);
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

  const devLogin = useCallback(() => {
    const header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }));
    const payload = btoa(JSON.stringify({ userID: 1, username: "admin" }));
    const mockToken = `${header}.${payload}.dev_signature`;
    localStorage.setItem(
      TOKEN_KEY,
      JSON.stringify({ access_token: mockToken, refresh_token: mockToken }),
    );
    localStorage.setItem(USER_KEY, JSON.stringify(1));
    localStorage.setItem("dev_mode", "true");
    emit();
  }, []);

  const logout = useCallback(async () => {
    const t = readTokens();
    if (t) {
      try {
        await authApi.logout(t.access_token, t.refresh_token);
      } catch {
        // 即使后端调用失败也清除本地状态
      }
    }
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    localStorage.removeItem("dev_mode");
    emit();
  }, []);

  return {
    tokens,
    userId,
    isAuthenticated,
    login,
    register,
    logout,
    devLogin,
    saveSession,
    setUserId,
  };
}
