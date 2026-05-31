import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

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
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary-50 via-white to-orange-50 p-4">
      <Card className="w-full max-w-md shadow-lg">
        <CardHeader className="space-y-1 text-center">
          <CardTitle className="text-3xl font-bold bg-gradient-to-r from-primary-500 to-primary-600 bg-clip-text text-transparent">
            创建账号
          </CardTitle>
          <CardDescription className="text-base">加入 NovelFlow 开始您的创作之旅</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">用户名 *</Label>
              <Input
                id="username"
                name="username"
                value={form.username}
                onChange={handleChange}
                required
                minLength={3}
                placeholder="至少3个字符"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">密码 *</Label>
              <Input
                id="password"
                type="password"
                name="password"
                value={form.password}
                onChange={handleChange}
                required
                minLength={6}
                placeholder="至少6个字符"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">邮箱</Label>
              <Input
                id="email"
                name="email"
                type="email"
                value={form.email}
                onChange={handleChange}
                placeholder="可选"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="nickname">昵称</Label>
              <Input
                id="nickname"
                name="nickname"
                value={form.nickname}
                onChange={handleChange}
                placeholder="可选"
              />
            </div>
            {error && (
              <p className="text-sm text-red-600 text-center bg-red-50 p-2 rounded-md">
                {error}
              </p>
            )}
            <Button className="w-full" type="submit" disabled={submitting}>
              {submitting ? "注册中..." : "注册"}
            </Button>
          </form>
          <div className="mt-4 text-center text-sm text-gray-600">
            已有账号？
            <Link to="/login" className="text-primary-500 hover:text-primary-600 font-medium ml-1">
              登录
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
