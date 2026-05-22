import client from "./client";
import type { ApiResponse, UpdateUserRequest, UserResponse } from "./types";

export async function getUser(id: number) {
  const { data } = await client.get<ApiResponse<UserResponse>>(`/users/${id}`);
  return data;
}

export async function updateUser(id: number, req: UpdateUserRequest) {
  const { data } = await client.put<ApiResponse<UserResponse>>(`/users/${id}`, req);
  return data;
}

export async function deleteUser(id: number) {
  const { data } = await client.delete<ApiResponse<null>>(`/users/${id}`);
  return data;
}
