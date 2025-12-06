import { Navigate, Route, Routes } from "react-router-dom";
import ProxiesPage from "./pages/Proxies";
import AuthPage from "./pages/Auth";
import AuthGuard from "./components/AuthGuard";
import HistoryPage from "./pages/History";

function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/proxies" />} />
      <Route path="/auth" element={<AuthPage />} />
      <Route
        path="/proxies"
        element={
          <AuthGuard>
            <ProxiesPage />
          </AuthGuard>
        }
      />
      <Route
        path="/history"
        element={
          <AuthGuard>
            <HistoryPage />
          </AuthGuard>
        }
      />
    </Routes>
  );
}

export default App;
