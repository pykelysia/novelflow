import { useCallback, useState } from "react";
import type {
  ChapterResult,
  GenerateRequest,
  GenerateStatusResponse,
  TaskItem,
} from "../api/types";
import * as generateApi from "../api/generate";

export function useGenerate() {
  const [tasks, setTasks] = useState<TaskItem[]>([]);
  const [loading, setLoading] = useState(false);

  const start = useCallback(async (req: GenerateRequest) => {
    const res = await generateApi.startGeneration(req);
    if (!res.data) throw new Error("start generation failed");
    return res.data;
  }, []);

  const fetchTasks = useCallback(async () => {
    setLoading(true);
    try {
      const res = await generateApi.listTasks();
      if (res.data) setTasks(res.data.tasks);
    } finally {
      setLoading(false);
    }
  }, []);

  const getStatus = useCallback(
    async (sessionId: string): Promise<GenerateStatusResponse> => {
      const res = await generateApi.getGenerationStatus(sessionId);
      if (!res.data) throw new Error("task not found");
      return res.data;
    },
    [],
  );

  const getResult = useCallback(
    async (sessionId: string): Promise<ChapterResult[]> => {
      const res = await generateApi.getGenerationResult(sessionId);
      return res.data?.chapters ?? [];
    },
    [],
  );

  return { tasks, loading, start, fetchTasks, getStatus, getResult };
}
