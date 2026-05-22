import axios, { AxiosError, InternalAxiosRequestConfig } from "axios";
import type { ApiResponse } from "./types";

const client = axios.create({
  baseURL: "/",
  timeout: 30_000,
});

function getTokens() {
  try {
    const raw = localStorage.getItem("tokens");
    return raw ? (JSON.parse(raw) as { access_token: string; refresh_token: string }) : null;
  } catch {
    return null;
  }
}

// 请求拦截器：自动附带 access_token
client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const tokens = getTokens();
  if (tokens?.access_token) {
    config.headers.Authorization = `Bearer ${tokens.access_token}`;
  }
  return config;
});

let isRefreshing = false;
let refreshSubscribers: ((token: string) => void)[] = [];

function onRefreshed(token: string) {
  refreshSubscribers.forEach((cb) => cb(token));
  refreshSubscribers = [];
}

// 响应拦截器：401 自动尝试 refresh
client.interceptors.response.use(
  (res) => res,
  async (error: AxiosError<ApiResponse>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };
    if (!originalRequest || originalRequest._retry) {
      return Promise.reject(error);
    }

    if (error.response?.status === 401 && error.config?.url !== "/auth/login") {
      // 开发模式下不尝试 refresh token，避免重定向到登录页
      if (localStorage.getItem("dev_mode")) {
        return Promise.reject(error);
      }
      const tokens = getTokens();
      if (!tokens?.refresh_token) {
        localStorage.removeItem("tokens");
        window.location.href = "/login";
        return Promise.reject(error);
      }

      if (isRefreshing) {
        return new Promise((resolve) => {
          refreshSubscribers.push((token: string) => {
            originalRequest.headers.Authorization = `Bearer ${token}`;
            resolve(client(originalRequest));
          });
        });
      }

      isRefreshing = true;
      originalRequest._retry = true;

      try {
        const { data } = await axios.post<ApiResponse<Token>>("/auth/refresh", {
          refresh_token: tokens.refresh_token,
        });

        const newTokens = data.data!;
        localStorage.setItem("tokens", JSON.stringify(newTokens));
        onRefreshed(newTokens.access_token);
        originalRequest.headers.Authorization = `Bearer ${newTokens.access_token}`;
        return client(originalRequest);
      } catch {
        localStorage.removeItem("tokens");
        window.location.href = "/login";
        return Promise.reject(error);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  },
);

interface Token {
  access_token: string;
  refresh_token: string;
}

export default client;
