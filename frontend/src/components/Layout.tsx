import { Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { Button } from "@/components/ui/button";

export default function Layout() {
  const { logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate("/login");
  };

  return (
    <div className="min-h-screen flex flex-col bg-gray-50">
      <header className="bg-gradient-to-r from-primary-500 to-primary-600 text-white shadow-lg">
        <div className="container mx-auto px-6 py-4 flex justify-between items-center">
          <h1
            onClick={() => navigate("/")}
            className="text-2xl font-bold cursor-pointer hover:opacity-90 transition-opacity"
          >
            NovelFlow
          </h1>
          <nav className="flex gap-3">
            <Button
              variant="ghost"
              onClick={() => navigate("/tasks")}
              className="text-white hover:bg-white/20"
            >
              任务
            </Button>
            <Button
              variant="ghost"
              onClick={() => navigate("/")}
              className="text-white hover:bg-white/20"
            >
              控制台
            </Button>
            <Button
              variant="ghost"
              onClick={handleLogout}
              className="text-white hover:bg-white/20"
            >
              退出登录
            </Button>
          </nav>
        </div>
      </header>
      {import.meta.env.DEV && (
        <div className="bg-amber-50 border-b border-amber-200 px-6 py-2">
          <div className="container mx-auto flex items-center gap-3 text-sm">
            <span className="font-semibold text-amber-800">Dev:</span>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate("/")}
              className="h-7 text-xs hover:bg-amber-100"
            >
              Dashboard
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate("/tasks/demo-dev-session")}
              className="h-7 text-xs hover:bg-amber-100"
            >
              TaskDetail
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate("/login")}
              className="h-7 text-xs hover:bg-amber-100"
            >
              Login
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate("/register")}
              className="h-7 text-xs hover:bg-amber-100"
            >
              Register
            </Button>
          </div>
        </div>
      )}
      <main className="flex-1 flex flex-col">
        <Outlet />
      </main>
    </div>
  );
}
