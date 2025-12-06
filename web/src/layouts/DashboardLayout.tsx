import { useState, type ReactNode } from "react";
import { NavLink, useNavigate } from "react-router-dom";

export function DashboardLayout({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();

  const logout = () => {
    localStorage.removeItem("token");
    navigate("/auth", { replace: true });
  };

  return (
    <div className="flex h-screen bg-darkBg text-darkText">
      <div className="lg:hidden fixed top-0 left-0 w-full bg-gray-800 p-3 flex items-center">
        <button onClick={() => setOpen(true)} className="text-white">
          <svg width="24" height="24" fill="currentColor" stroke="currentColor">
            <path d="M3 6h18M3 12h18M3 18h18" />
          </svg>
        </button>
        <span className="ml-4 font-semibold">Menu</span>
      </div>

      <aside
        className={`
          fixed top-0 left-0 h-full w-64 bg-gray-800 p-6 transform transition-transform duration-200
          lg:static lg:translate-x-0 
          ${open ? "translate-x-0" : "-translate-x-full"}
        `}
      >
        <button
          className="lg:hidden mb-4 text-white"
          onClick={() => setOpen(false)}
        >
          <svg width="24" height="24" fill="currentColor" stroke="currentColor">
            <path d="M6 6l12 12M6 18L18 6" />
          </svg>
        </button>

        <h2 className="text-xl font-bold mb-4">Proxy Dashboard</h2>

        <nav className="flex flex-col gap-2">
          <NavLink
            to="/proxies"
            className={({ isActive }) =>
              isActive ? "underline" : "text-gray-300"
            }
            onClick={() => setOpen(false)}
          >
            Proxies
          </NavLink>

          <NavLink
            to="/history"
            className={({ isActive }) =>
              isActive ? "underline" : "text-gray-300"
            }
            onClick={() => setOpen(false)}
          >
            History
          </NavLink>

          <button
            onClick={logout}
            className="text-left text-red-400 hover:text-red-300 mt-4"
          >
            Logout
          </button>
        </nav>
      </aside>

      {open && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 lg:hidden"
          onClick={() => setOpen(false)}
        ></div>
      )}

      <main className="flex-1 overflow-auto p-6 pt-14 lg:pt-6">{children}</main>
    </div>
  );
}
