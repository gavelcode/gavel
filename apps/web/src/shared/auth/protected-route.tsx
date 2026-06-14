import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/entities/user/use-auth";
import { Spinner } from "@/shared/ui/spinner";

export function ProtectedRoute() {
  const { user, loading } = useAuth();

  if (loading) return <Spinner />;
  if (!user) return <Navigate to="/login" replace />;

  return <Outlet />;
}
