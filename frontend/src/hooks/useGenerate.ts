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
    } catch {
      // 开发模式下，API 401 时使用 mock 数据展示 UI
      if (import.meta.env.DEV) {
        setTasks([
          { session_id: "demo-dev-session", status: "completed", created_at: "2026-05-22 10:00", updated_at: "2026-05-22 10:30" },
          { session_id: "demo-dev-running", status: "running", created_at: "2026-05-22 11:00", updated_at: "2026-05-22 11:05" },
        ]);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  const getStatus = useCallback(
    async (sessionId: string): Promise<GenerateStatusResponse> => {
      // 开发模式下 mock 数据
      if (import.meta.env.DEV && sessionId.startsWith("demo-dev-")) {
        return {
          session_id: sessionId,
          status: sessionId === "demo-dev-running" ? "running" : "completed",
          created_at: "2026-05-22 10:00",
          updated_at: "2026-05-22 10:30",
        };
      }
      const res = await generateApi.getGenerationStatus(sessionId);
      if (!res.data) throw new Error("task not found");
      return res.data;
    },
    [],
  );

  const getResult = useCallback(
    async (sessionId: string): Promise<ChapterResult[]> => {
      // 开发模式下 mock 数据
      if (import.meta.env.DEV && sessionId.startsWith("demo-dev-")) {
        return [
          { title: "第一章 觉醒", content: "林玄睁开眼，发现自己正躺在一片荒芜的山坡上。\n\n天空是诡异的暗红色，空气中弥漫着硫磺的气息。他最后的记忆停留在实验室爆炸的那一瞬间——量子纠缠实验出了致命错误。\n\n\"你醒了。\"一个淡漠的声音从身后传来。\n\n林玄猛地翻身而起，看到一个白发老者正盘腿坐在三米外的一块青石上。老者的服饰如同古画中走出的仙人，但眼神却像激光一样锐利。\n\n\"这里是哪里？\"林玄警惕地问。\n\n\"天玄大陆，落霞山脉。\"老者缓缓说道，\"你从界隙中坠落，若非老夫恰好路过，你已被妖兽啃食殆尽。\"" },
          { title: "第二章 异界法则", content: "接下来的三天，林玄逐渐接受了穿越的事实。\n\n他体内残存的量子能量与此界的灵气产生了奇异的共鸣，形成了一种前所未有的力量——他将其命名为'量子灵力'。\n\n白发老者自称'青云真人'，是附近青云宗的长老。他发现林玄的体质特殊，有意收其为徒。\n\n\"你体内的力量很古怪，\"青云真人皱眉道，\"既有天地灵气的波动，又带着某种……规则之外的气息。\"" },
        ];
      }
      const res = await generateApi.getGenerationResult(sessionId);
      return res.data?.chapters ?? [];
    },
    [],
  );

  return { tasks, loading, start, fetchTasks, getStatus, getResult };
}
