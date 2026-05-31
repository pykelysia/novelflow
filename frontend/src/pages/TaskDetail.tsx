import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import type { ChapterResult } from "../api/types";
import { useGenerate } from "../hooks/useGenerate";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import { Separator } from "@/components/ui/separator";

export default function TaskDetail() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const { getStatus, getResult } = useGenerate();
  const [chapters, setChapters] = useState<ChapterResult[]>([]);
  const [status, setStatus] = useState<string>("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    if (!sessionId) return;
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
    fetchData();
  }, [fetchData]);

  // 如果还在运行，轮询
  useEffect(() => {
    if (!sessionId || status === "completed" || status === "failed" || !status) return;
    const timer = setInterval(fetchData, 3000);
    return () => clearInterval(timer);
  }, [sessionId, status, fetchData]);

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-10 w-32" />
        <Card>
          <CardContent className="p-6">
            <Skeleton className="h-6 w-48 mb-4" />
            <Skeleton className="h-4 w-full mb-2" />
            <Skeleton className="h-4 w-full mb-2" />
            <Skeleton className="h-4 w-3/4" />
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="outline" onClick={() => navigate("/")}>
          ← 返回
        </Button>
        <span className="text-sm text-gray-600">会话: {sessionId}</span>
      </div>

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
              <p className="text-sm text-gray-600">
                状态: {status === "running" ? "生成中" : "等待中"}
              </p>
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
