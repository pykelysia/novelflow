import client from "./client";
import type {
  ApiResponse,
  LoginRequest,
  RegisterRequest,
  TokenResponse,
  UserResponse,
} from "./types";

export async function login(req: LoginRequest) {
  const { data } = await client.post<ApiResponse<TokenResponse>>("/auth/login", req);
  return data;
}

export async function register(req: RegisterRequest) {
  const { data } = await client.post<ApiResponse<UserResponse>>("/auth/register", req);
  return data;
}

export async function refresh(refreshToken: string) {
  const { data } = await client.post<ApiResponse<TokenResponse>>("/auth/refresh", {
    refresh_token: refreshToken,
  });
  return data;
}

export async function logout(accessToken: string, refreshToken: string) {
  const { data } = await client.post<ApiResponse<null>>("/auth/logout", {
    access_token: accessToken,
    refresh_token: refreshToken,
  });
  return data;
}
