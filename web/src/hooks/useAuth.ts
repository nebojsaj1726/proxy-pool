import { useMutation } from "@tanstack/react-query";
import { api } from "../api/api";
import { useNavigate } from "react-router-dom";

interface AuthCredentials {
  username: string;
  password: string;
}

export function useAuth() {
  const navigate = useNavigate();

  const loginMutation = useMutation({
    mutationFn: ({ username, password }: AuthCredentials) =>
      api.login(username, password),
    onSuccess: (data) => {
      localStorage.setItem("token", data.token);
      navigate("/proxies", { replace: true });
    },
  });

  const registerMutation = useMutation({
    mutationFn: ({ username, password }: AuthCredentials) =>
      api.register(username, password),
    onSuccess: async (_data, variables) => {
      await loginMutation.mutateAsync(variables);
    },
  });

  const resetError = () => {
    loginMutation.reset();
    registerMutation.reset();
  };

  return {
    login: loginMutation.mutateAsync,
    register: registerMutation.mutateAsync,
    loading: loginMutation.isPending || registerMutation.isPending,
    error: loginMutation.error || registerMutation.error,
    resetError,
  };
}
