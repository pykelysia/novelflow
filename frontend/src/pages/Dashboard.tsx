import { FormEvent, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { useGenerate } from "../hooks/useGenerate";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";

const GENRE_OPTIONS = [
  "玄幻",
  "仙侠",
  "都市",
  "历史",
  "科幻",
  "悬疑",
  "言情",
  "轻小说",
];

export default function Dashboard() {
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const { tasks, loading, start, fetchTasks } = useGenerate();

  const [genre, setGenre] = useState("");
  const [concept, setConcept] = useState("");
  const [protagonist, setProtagonist] = useState("");
  const [worldSetting, setWorldSetting] = useState("");
  const [chapterCount, setChapterCount] = useState(3);
  const [style, setStyle] = useState("");
  const [requirements, setRequirements] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!isAuthenticated) return;
    fetchTasks();
  }, [isAuthenticated, fetchTasks]);

  // 自动轮询运行中的任务
  useEffect(() => {
    if (!isAuthenticated) return;
    const hasRunning = tasks.some((t) => t.status === "running" || t.status === "pending");
    if (!hasRunning) return;
    const timer = setInterval(fetchTasks, 5000);
    return () => clearInterval(timer);
  }, [isAuthenticated, tasks, fetchTasks]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!genre || !concept) {
      setError("请填写题材和创意概念");
      return;
    }
    setError("");
    setSubmitting(true);
    try {
      await start({ genre, concept, protagonist: protagonist || undefined, world_setting: worldSetting || undefined, chapter_count: chapterCount, style: style || undefined, requirements: requirements || undefined });
      setConcept("");
      setProtagonist("");
      setWorldSetting("");
      setStyle("");
      setRequirements("");
      await fetchTasks();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "创建任务失败";
      setError(msg);
    } finally {
      setSubmitting(false);
    }
  };

  const getStatusVariant = (status: string) => {
    switch (status) {
      case "pending": return "pending";
      case "running": return "running";
      case "completed": return "completed";
      default: return "failed";
    }
  };

  const statusLabel = (s: string) => {
    switch (s) {
      case "pending": return "等待中";
      case "running": return "生成中";
      case "completed": return "已完成";
      default: return "失败";
    }
  };

  return (
    <div className="space-y-6">
      {import.meta.env.DEV && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 text-sm text-blue-800">
          调试: auth={isAuthenticated ? 'OK' : 'NO'} | tasks={tasks.length} | loading={loading ? 'Y' : 'N'} | dev_mode={!!localStorage.getItem('dev_mode')}
        </div>
      )}

      <Card className="shadow-md">
        <CardHeader>
          <CardTitle className="text-2xl">新建生成任务</CardTitle>
          <CardDescription>填写小说创意，让 AI 为您创作精彩内容</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="genre">题材 *</Label>
                <Select id="genre" value={genre} onChange={(e) => setGenre(e.target.value)} required>
                  <option value="">请选择</option>
                  {GENRE_OPTIONS.map((g) => (
                    <option key={g} value={g}>{g}</option>
                  ))}
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="chapterCount">章节数量</Label>
                <Input
                  id="chapterCount"
                  type="number"
                  min={1}
                  max={50}
                  value={chapterCount}
                  onChange={(e) => setChapterCount(Number(e.target.value))}
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="concept">创意概念 *</Label>
              <Textarea
                id="concept"
                value={concept}
                onChange={(e) => setConcept(e.target.value)}
                placeholder="描述你的小说核心创意..."
                required
                rows={3}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="protagonist">主角设定</Label>
              <Input
                id="protagonist"
                value={protagonist}
                onChange={(e) => setProtagonist(e.target.value)}
                placeholder="主角姓名、性格、能力等"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="worldSetting">世界观设定</Label>
              <Textarea
                id="worldSetting"
                value={worldSetting}
                onChange={(e) => setWorldSetting(e.target.value)}
                placeholder="故事发生的世界背景..."
                rows={3}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="style">风格要求</Label>
              <Input
                id="style"
                value={style}
                onChange={(e) => setStyle(e.target.value)}
                placeholder="文风、节奏等要求"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="requirements">其他要求</Label>
              <Textarea
                id="requirements"
                value={requirements}
                onChange={(e) => setRequirements(e.target.value)}
                placeholder="其他需要 AI 注意的事项..."
                rows={2}
              />
            </div>

            {error && (
              <p className="text-sm text-red-600 bg-red-50 p-3 rounded-md">
                {error}
              </p>
            )}

            <Button className="w-full" type="submit" disabled={submitting}>
              {submitting ? "提交中..." : "开始生成"}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card className="shadow-md">
        <CardHeader>
          <CardTitle className="text-2xl flex items-center gap-2">
            任务列表
            {loading && <span className="text-sm font-normal text-gray-500">(刷新中...)</span>}
          </CardTitle>
          <CardDescription>查看您的创作任务进度</CardDescription>
        </CardHeader>
        <CardContent>
          {loading && tasks.length === 0 ? (
            <div className="space-y-3">
              <Skeleton className="h-20 w-full" />
              <Skeleton className="h-20 w-full" />
            </div>
          ) : tasks.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-gray-500 text-sm">暂无任务</p>
              <p className="text-gray-400 text-xs mt-1">创建您的第一个小说生成任务吧！</p>
            </div>
          ) : (
            <div className="space-y-3">
              {tasks.map((t) => (
                <Card key={t.session_id} className="hover:shadow-md transition-shadow">
                  <CardContent className="p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <Badge variant={getStatusVariant(t.status) as any}>
                          {statusLabel(t.status)}
                        </Badge>
                        <span className="text-sm text-gray-600">
                          {t.created_at}
                        </span>
                      </div>
                      <div>
                        {t.status === "completed" ? (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => navigate(`/tasks/${t.session_id}`)}
                            className="border-green-200 text-green-700 hover:bg-green-50"
                          >
                            查看
                          </Button>
                        ) : t.status === "failed" ? (
                          <span className="text-sm text-red-600">{t.error || "失败"}</span>
                        ) : null}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
