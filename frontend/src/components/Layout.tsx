import { Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";

export default function Layout() {
  const { logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate("/login");
  };

  return (
    <div className="layout">
      <header className="header">
        <h1 onClick={() => navigate("/")} style={{ cursor: "pointer" }}>
          NovelFlow
        </h1>
        <nav>
          <button onClick={() => navigate("/")}>控制台</button>
          <button onClick={handleLogout}>退出登录</button>
        </nav>
      </header>
      {import.meta.env.DEV && (
        <div className="dev-toolbar">
          <span className="dev-toolbar-label">Dev:</span>
          <button onClick={() => navigate("/")}>Dashboard</button>
          <button onClick={() => navigate("/tasks/demo-dev-session")}>TaskDetail</button>
          <button onClick={() => navigate("/login")}>Login</button>
          <button onClick={() => navigate("/register")}>Register</button>
        </div>
      )}
      <main className="main">
        <Outlet />
      </main>
    </div>
  );
}
