import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import type { ChapterResult } from "../api/types";
import { useGenerate } from "../hooks/useGenerate";

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

  if (loading) return <div className="card"><p>加载中...</p></div>;

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <button className="btn" onClick={() => navigate("/")} style={{ border: "1px solid #ddd" }}>
          &larr; 返回
        </button>
        <span style={{ marginLeft: 12, fontSize: 14, color: "#666" }}>
          会话: {sessionId}
        </span>
      </div>

      {error && (
        <div className="card" style={{ borderLeft: "4px solid #d32f2f" }}>
          <p style={{ color: "#d32f2f" }}>{error}</p>
        </div>
      )}

      {status === "running" || status === "pending" ? (
        <div className="card" style={{ textAlign: "center" }}>
          <p>AI 正在创作中，请稍候...</p>
          <p style={{ fontSize: 13, color: "#888", marginTop: 8 }}>状态: {status === "running" ? "生成中" : "等待中"}</p>
        </div>
      ) : null}

      {chapters.length > 0 && (
        <div className="card">
          <h3>生成结果（共 {chapters.length} 章）</h3>
          {chapters.map((ch, i) => (
            <div key={i} className="chapter">
              <h4>{ch.title || `第 ${i + 1} 章`}</h4>
              <pre>{ch.content}</pre>
            </div>
          ))}
        </div>
      )}

      {status === "completed" && chapters.length === 0 && !error && (
        <div className="card"><p>暂无内容</p></div>
      )}
    </div>
  );
}
