import { FormEvent, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { useGenerate } from "../hooks/useGenerate";

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
  const { isAuthenticated, tokens } = useAuth();
  const navigate = useNavigate();
  const { tasks, loading, start, fetchTasks } = useGenerate();

  console.log("[Dev] Dashboard render", { isAuthenticated, hasToken: !!tokens, tasks: tasks.length, loading });

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

  const statusClass = (s: string) => {
    switch (s) {
      case "pending": return "status-pending";
      case "running": return "status-running";
      case "completed": return "status-completed";
      default: return "status-failed";
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
    <div className="dashboard">
      {import.meta.env.DEV && (
        <div className="dev-badge" style={{ background: '#e3f2fd', border: '1px solid #1976d2', color: '#1565c0' }}>
          调试: auth={isAuthenticated ? 'OK' : 'NO'}  tasks={tasks.length}  loading={loading ? 'Y' : 'N'}  dev_mode={!!localStorage.getItem('dev_mode')}
        </div>
      )}
      {import.meta.env.DEV && (
        <div className="dev-badge">开发模式 · 显示 mock 数据</div>
      )}
      <div className="card">
        <h3>新建生成任务</h3>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>题材 *</label>
            <select value={genre} onChange={(e) => setGenre(e.target.value)} required>
              <option value="">请选择</option>
              {GENRE_OPTIONS.map((g) => (
                <option key={g} value={g}>{g}</option>
              ))}
            </select>
          </div>
          <div className="form-group">
            <label>创意概念 *</label>
            <textarea
              value={concept}
              onChange={(e) => setConcept(e.target.value)}
              placeholder="描述你的小说核心创意..."
              required
            />
          </div>
          <div className="form-group">
            <label>主角设定</label>
            <input
              value={protagonist}
              onChange={(e) => setProtagonist(e.target.value)}
              placeholder="主角姓名、性格、能力等"
            />
          </div>
          <div className="form-group">
            <label>世界观设定</label>
            <textarea
              value={worldSetting}
              onChange={(e) => setWorldSetting(e.target.value)}
              placeholder="故事发生的世界背景..."
            />
          </div>
          <div className="form-group">
            <label>章节数量</label>
            <input
              type="number"
              min={1}
              max={50}
              value={chapterCount}
              onChange={(e) => setChapterCount(Number(e.target.value))}
            />
          </div>
          <div className="form-group">
            <label>风格要求</label>
            <input
              value={style}
              onChange={(e) => setStyle(e.target.value)}
              placeholder="文风、节奏等要求"
            />
          </div>
          <div className="form-group">
            <label>其他要求</label>
            <textarea
              value={requirements}
              onChange={(e) => setRequirements(e.target.value)}
              placeholder="其他需要 AI 注意的事项..."
            />
          </div>
          {error && <p className="error-msg">{error}</p>}
          <button className="btn btn-primary" type="submit" disabled={submitting}>
            {submitting ? "提交中..." : "开始生成"}
          </button>
        </form>
      </div>

      <div className="card">
        <h3>任务列表 {loading && "(刷新中...)"}</h3>
        {tasks.length === 0 ? (
          <p style={{ color: "#888", fontSize: 14 }}>暂无任务</p>
        ) : (
          <ul className="task-list">
            {tasks.map((t) => (
              <li key={t.session_id} className="task-item">
                <div>
                  <span className={`status ${statusClass(t.status)}`}>
                    {statusLabel(t.status)}
                  </span>
                  <span style={{ marginLeft: 8, fontSize: 13, color: "#666" }}>
                    {t.created_at}
                  </span>
                </div>
                <div>
                  {t.status === "completed" ? (
                    <button
                      className="btn"
                      style={{ background: "#e8f5e9", color: "#2e7d32" }}
                      onClick={() => navigate(`/tasks/${t.session_id}`)}
                    >
                      查看
                    </button>
                  ) : t.status === "failed" ? (
                    <span style={{ fontSize: 13, color: "#c62828" }}>{t.error || "失败"}</span>
                  ) : null}
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
