import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";

export default function Login() {
  const { login, setUserId } = useAuth();
  const navigate = useNavigate();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setSubmitting(true);
    try {
      const tokens = await login(username, password);
      // 解析用户 ID — 后端 JWT 的 userID claim 无法直接获取，调一次 /users 不方便
      // 这里通过简单的 claims 解析获取（生产环境应有 /auth/me 接口）
      // 先用 tokens 做后续请求，userID 可以从 /generate 等接口间接获取
      // 临时方案：检查 localStorage 中的 userID
      let uid = localStorage.getItem("current_user");
      if (!uid) {
        // 用 access_token 调用一个受保护接口拿到用户 id/信息
        // 先默认从 token 的 payload 中解析（base64 decode）
        try {
          const payload = JSON.parse(atob(tokens.access_token.split(".")[1]));
          if (payload.userID) {
            setUserId(payload.userID);
          }
        } catch {
          // 解析失败不影响登录
        }
      }
      navigate("/");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "登录失败";
      setError(msg);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="form-card">
      <h2>登录</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>用户名</label>
          <input
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
            autoFocus
          />
        </div>
        <div className="form-group">
          <label>密码</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>
        {error && <p className="error-msg">{error}</p>}
        <button className="btn btn-primary" type="submit" disabled={submitting}>
          {submitting ? "登录中..." : "登录"}
        </button>
      </form>
      <div className="link-row">
        还没有账号？<Link to="/register">注册</Link>
      </div>
    </div>
  );
}
