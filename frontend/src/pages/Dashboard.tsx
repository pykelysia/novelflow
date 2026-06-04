import { FormEvent, useState } from "react";
import { useGenerate } from "../hooks/useGenerate";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select } from "@/components/ui/select";
import { Button } from "@/components/ui/button";

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
  const { start } = useGenerate();

  const [genre, setGenre] = useState("");
  const [concept, setConcept] = useState("");
  const [protagonist, setProtagonist] = useState("");
  const [worldSetting, setWorldSetting] = useState("");
  const [chapterCount, setChapterCount] = useState(3);
  const [style, setStyle] = useState("");
  const [requirements, setRequirements] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

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
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "创建任务失败";
      setError(msg);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="container mx-auto px-6 py-8 max-w-3xl">
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
    </div>
  );
}
