import { useAuth } from "../hooks/useAuth";
import { useEffect, useRef, useState } from "react";
import { Button } from "../components/Button";
import { useNavigate } from "react-router-dom";

export default function AuthPage() {
  const { login, register, error, loading, resetError } = useAuth();

  const usernameRef = useRef<HTMLInputElement>(null);

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const [isRegister, setIsRegister] = useState(false);

  const navigate = useNavigate();
  const token = localStorage.getItem("token");

  useEffect(() => {
    if (token) {
      navigate("/proxies", { replace: true });
    }
  }, [token, navigate]);

  useEffect(() => {
    usernameRef.current?.focus();
  }, []);

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    resetError();

    const credentials = { username, password };

    if (isRegister) {
      register(credentials);
    } else {
      login(credentials);
    }
  };

  const toggleMode = () => {
    setIsRegister((prev) => !prev);
    setUsername("");
    setPassword("");
    resetError();
  };

  return (
    <div className="flex items-center justify-center h-screen bg-darkBg text-darkText">
      <form
        onSubmit={onSubmit}
        className="bg-gray-800 p-8 rounded-lg w-96 shadow-xl"
      >
        <h1 className="text-2xl font-bold mb-6 text-center">
          {isRegister ? "Register" : "Login"}
        </h1>

        <div className="mb-4">
          <label className="block text-sm mb-1">Username</label>
          <input
            ref={usernameRef}
            className="w-full p-2 bg-gray-700 rounded text-white outline-none focus:ring-2 focus:ring-blue-500 autofill:bg-gray-700"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            autoComplete="username"
          />
        </div>

        <div className="mb-6">
          <label className="block text-sm mb-1">Password</label>
          <input
            type="password"
            className="w-full p-2 bg-gray-700 rounded text-white outline-none focus:ring-2 focus:ring-blue-500 autofill:bg-gray-700"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
          />
        </div>

        {error && (
          <p className="text-red-400 text-sm mb-4 text-center">
            {error.message}
          </p>
        )}

        <Button type="submit" disabled={loading} className="w-full py-2.5">
          {loading ? "Please wait..." : isRegister ? "Create Account" : "Login"}
        </Button>

        <button
          type="button"
          onClick={toggleMode}
          className="mt-4 w-full text-sm text-center text-gray-300 hover:text-white"
        >
          {isRegister
            ? "Already have an account? Login"
            : "Need an account? Register"}
        </button>
      </form>
    </div>
  );
}
