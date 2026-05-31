import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

export default function Login() {
  const { login, setUserId, devLogin } = useAuth();
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
      let uid = localStorage.getItem("current_user");
      if (!uid) {
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

  const handleDevLogin = () => {
    devLogin();
    navigate("/");
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary-50 via-white to-orange-50 p-4">
      <Card className="w-full max-w-md shadow-lg">
        <CardHeader className="space-y-1 text-center">
          <CardTitle className="text-3xl font-bold bg-gradient-to-r from-primary-500 to-primary-600 bg-clip-text text-transparent">
            NovelFlow
          </CardTitle>
          <CardDescription className="text-base">登录您的账号开始创作</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">用户名</Label>
              <Input
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                autoFocus
                placeholder="请输入用户名"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">密码</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                placeholder="请输入密码"
              />
            </div>
            {error && (
              <p className="text-sm text-red-600 text-center bg-red-50 p-2 rounded-md">
                {error}
              </p>
            )}
            <Button className="w-full" type="submit" disabled={submitting}>
              {submitting ? "登录中..." : "登录"}
            </Button>
          </form>
          <div className="mt-4 text-center text-sm text-gray-600">
            还没有账号？
            <Link to="/register" className="text-primary-500 hover:text-primary-600 font-medium ml-1">
              注册
            </Link>
          </div>
          {import.meta.env.DEV && (
            <div className="mt-6 pt-6 border-t border-gray-200">
              <p className="text-xs text-gray-500 text-center mb-3">开发模式</p>
              <Button
                variant="secondary"
                className="w-full"
                type="button"
                onClick={handleDevLogin}
              >
                Admin 开发者入口
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
