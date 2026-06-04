import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useGenerate } from "../hooks/useGenerate";
import type { ChapterResult, TaskItem } from "../api/types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Separator } from "@/components/ui/separator";
import { Alert, AlertDescription } from "@/components/ui/alert";

function statusLabel(s: string) {
  switch (s) {
    case "pending": return "等待中";
    case "running": return "生成中";
    case "completed": return "已完成";
    default: return "失败";
  }
}

function getStatusVariant(status: string) {
  switch (status) {
    case "pending": return "pending";
    case "running": return "running";
    case "completed": return "completed";
    default: return "failed";
  }
}

function TaskDetailPanel({ sessionId }: { sessionId: string }) {
  const { getStatus, getResult } = useGenerate();
  const [chapters, setChapters] = useState<ChapterResult[]>([]);
  const [status, setStatus] = useState<string>("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    try {
      const s = await getStatus(sessionId);
      setStatus(s.status);
      if (s.status === "completed") {
        const chs = await getResult(sessionId);
        setChapters(chs);
      } else if (s.status === "failed") {
        setError(s.error || "生成失败");
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "获取任务失败");
    } finally {
      setLoading(false);
    }
  }, [sessionId, getStatus, getResult]);

  useEffect(() => {
    setLoading(true);
    setChapters([]);
    setStatus("");
    setError("");
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    if (!sessionId || status === "completed" || status === "failed" || !status) return;
    const timer = setInterval(fetchData, 3000);
    return () => clearInterval(timer);
  }, [sessionId, status, fetchData]);

  if (loading) {
    return (
      <div className="p-6 space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-3/4" />
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6 overflow-y-auto">
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      {(status === "running" || status === "pending") && (
        <Card className="border-blue-200 bg-blue-50">
          <CardContent className="p-6 text-center">
            <div className="flex flex-col items-center gap-3">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500"></div>
              <p className="text-lg font-medium">AI 正在创作中，请稍候...</p>
              <p className="text-sm text-gray-600">状态: {statusLabel(status)}</p>
            </div>
          </CardContent>
        </Card>
      )}
      {chapters.length > 0 && (
        <Card className="shadow-md">
          <CardHeader>
            <CardTitle className="text-2xl">生成结果（共 {chapters.length} 章）</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {chapters.map((ch, i) => (
              <div key={i}>
                {i > 0 && <Separator className="my-6" />}
                <div className="space-y-3">
                  <h4 className="text-xl font-semibold text-gray-900">
                    {ch.title || `第 ${i + 1} 章`}
                  </h4>
                  <div className="bg-gray-50 rounded-lg p-6 border border-gray-200">
                    <pre className="whitespace-pre-wrap font-sans text-sm leading-relaxed text-gray-800">
                      {ch.content}
                    </pre>
                  </div>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
      {status === "completed" && chapters.length === 0 && !error && (
        <Card>
          <CardContent className="p-12 text-center">
            <p className="text-gray-500">暂无内容</p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

export default function TaskListPage() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const { tasks, loading, fetchTasks } = useGenerate();

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  useEffect(() => {
    const hasRunning = tasks.some((t) => t.status === "running" || t.status === "pending");
    if (!hasRunning) return;
    const timer = setInterval(fetchTasks, 5000);
    return () => clearInterval(timer);
  }, [tasks, fetchTasks]);

  return (
    <div className="flex flex-1 overflow-hidden">
      <aside className="w-72 flex-shrink-0 bg-white border-r border-gray-200 flex flex-col overflow-hidden">
        <div className="px-4 py-4 border-b border-gray-200">
          <h2 className="text-base font-semibold text-gray-900">任务列表</h2>
          {loading && <p className="text-xs text-gray-400 mt-0.5">刷新中...</p>}
        </div>
        <div className="flex-1 overflow-y-auto">
          {loading && tasks.length === 0 ? (
            <div className="p-4 space-y-3">
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
            </div>
          ) : tasks.length === 0 ? (
            <div className="p-6 text-center">
              <p className="text-sm text-gray-500">暂无任务</p>
            </div>
          ) : (
            <div className="py-2">
              {tasks.map((t: TaskItem) => (
                <button
                  key={t.session_id}
                  onClick={() => navigate(`/tasks/${t.session_id}`)}
                  className={`w-full text-left px-4 py-3 border-b border-gray-100 hover:bg-gray-50 transition-colors ${
                    sessionId === t.session_id ? "bg-primary-50 border-l-2 border-l-primary-500" : ""
                  }`}
                >
                  <div className="flex items-center gap-2 mb-1">
                    <Badge variant={getStatusVariant(t.status) as any}>
                      {statusLabel(t.status)}
                    </Badge>
                  </div>
                  <p className="text-xs text-gray-500 truncate">{t.created_at}</p>
                  <p className="text-xs text-gray-400 truncate font-mono">{t.session_id}</p>
                </button>
              ))}
            </div>
          )}
        </div>
      </aside>
      <main className="flex-1 overflow-y-auto bg-gray-50">
        {sessionId ? (
          <TaskDetailPanel sessionId={sessionId} />
        ) : (
          <div className="flex items-center justify-center h-full">
            <div className="text-center">
              <p className="text-gray-500">请从左侧选择一个任务</p>
              <p className="text-sm text-gray-400 mt-1">选中任务后可查看生成内容</p>
              <Button variant="outline" className="mt-4" onClick={() => navigate("/")}>
                去新建任务
              </Button>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
