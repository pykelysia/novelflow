import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";

export default function ProtectedRoute() {
  const { isAuthenticated } = useAuth();
  // 开发模式下，只要有 dev_mode 标记就直接放行
  if (import.meta.env.DEV && localStorage.getItem("dev_mode")) {
    return <Outlet />;
  }
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return <Outlet />;
}
