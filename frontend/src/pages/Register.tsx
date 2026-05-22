import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";

export default function Register() {
  const { register, login, setUserId } = useAuth();
  const navigate = useNavigate();
  const [form, setForm] = useState({
    username: "",
    password: "",
    email: "",
    nickname: "",
  });
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm((prev) => ({ ...prev, [e.target.name]: e.target.value }));
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setSubmitting(true);
    try {
      await register(form.username, form.password, form.email || undefined, form.nickname || undefined);
      // 注册成功后自动登录
      const tokens = await login(form.username, form.password);
      try {
        const payload = JSON.parse(atob(tokens.access_token.split(".")[1]));
        if (payload.userID) setUserId(payload.userID);
      } catch {
        // ignore
      }
      navigate("/");
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "注册失败";
      setError(msg);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="form-card">
      <h2>注册</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>用户名 *</label>
          <input name="username" value={form.username} onChange={handleChange} required minLength={3} />
        </div>
        <div className="form-group">
          <label>密码 *</label>
          <input
            type="password"
            name="password"
            value={form.password}
            onChange={handleChange}
            required
            minLength={6}
          />
        </div>
        <div className="form-group">
          <label>邮箱</label>
          <input name="email" type="email" value={form.email} onChange={handleChange} />
        </div>
        <div className="form-group">
          <label>昵称</label>
          <input name="nickname" value={form.nickname} onChange={handleChange} />
        </div>
        {error && <p className="error-msg">{error}</p>}
        <button className="btn btn-primary" type="submit" disabled={submitting}>
          {submitting ? "注册中..." : "注册"}
        </button>
      </form>
      <div className="link-row">
        已有账号？<Link to="/login">登录</Link>
      </div>
    </div>
  );
}
