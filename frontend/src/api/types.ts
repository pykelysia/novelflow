// ---- 统一响应 ----

export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data?: T;
}

// ---- 认证 ----

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
  email?: string;
  nickname?: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
}

export interface RefreshRequest {
  refresh_token: string;
}

export interface LogoutRequest {
  access_token: string;
  refresh_token: string;
}

// ---- 用户 ----

export interface UserResponse {
  id: number;
  username: string;
  email: string;
  nickname: string;
  avatar: string;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface UpdateUserRequest {
  email?: string;
  nickname?: string;
  avatar?: string;
  status?: number;
}

// ---- 生成 ----

export interface GenerateRequest {
  genre: string;
  concept: string;
  protagonist?: string;
  world_setting?: string;
  chapter_count?: number;
  style?: string;
  requirements?: string;
}

export interface GenerateResponse {
  session_id: string;
  status: string;
}

export interface GenerateStatusResponse {
  session_id: string;
  status: string;
  error?: string;
  created_at: string;
  updated_at: string;
}

export interface ChapterResult {
  title: string;
  content: string;
}

export interface GenerateResultResponse {
  session_id: string;
  status: string;
  chapters?: ChapterResult[];
  error?: string;
}

export interface TaskItem {
  session_id: string;
  status: string;
  error?: string;
  created_at: string;
  updated_at: string;
}

export interface ListTasksResponse {
  tasks: TaskItem[];
}
