import { type ReactNode, useEffect } from "react";
import { useNavigate } from "react-router-dom";

export default function AuthGuard({ children }: { children: ReactNode }) {
  const navigate = useNavigate();
  const token = localStorage.getItem("token");

  useEffect(() => {
    if (!token) {
      navigate("/auth", { replace: true });
    }
  }, [token, navigate]);

  if (!token) return null;
  return <>{children}</>;
}
