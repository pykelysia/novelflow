import client from "./client";
import type {
  ApiResponse,
  GenerateRequest,
  GenerateResponse,
  GenerateResultResponse,
  GenerateStatusResponse,
  ListTasksResponse,
} from "./types";

export async function startGeneration(req: GenerateRequest) {
  const { data } = await client.post<ApiResponse<GenerateResponse>>("/generate", req);
  return data;
}

export async function listTasks() {
  const { data } = await client.get<ApiResponse<ListTasksResponse>>("/generate/tasks");
  return data;
}

export async function getGenerationStatus(sessionId: string) {
  const { data } = await client.get<ApiResponse<GenerateStatusResponse>>(
    `/generate/${sessionId}`,
  );
  return data;
}

export async function getGenerationResult(sessionId: string) {
  const { data } = await client.get<ApiResponse<GenerateResultResponse>>(
    `/generate/${sessionId}/result`,
  );
  return data;
}
