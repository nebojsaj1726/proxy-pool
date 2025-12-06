import { DashboardLayout } from "../layouts/DashboardLayout";
import { Skeleton } from "../components/Skeleton";
import { useProxyStats } from "../hooks/useProxies";

export default function HistoryPage() {
  const { data: stats, isLoading, isError, error } = useProxyStats();

  return (
    <DashboardLayout>
      <h1 className="text-2xl font-bold mb-4 text-darkText">Proxy History</h1>

      {isLoading && <Skeleton className="h-64 w-full" />}

      {isError && (
        <div className="text-red-400">
          {error.message || "Failed to load stats"}
        </div>
      )}

      {stats && (
        <div className="overflow-auto rounded-lg border border-gray-700 bg-darkCard">
          <table className="w-full text-left text-darkText">
            <thead className="bg-gray-800">
              <tr>
                <th className="px-4 py-2">URL</th>
                <th className="px-4 py-2">Score</th>
                <th className="px-4 py-2">Usage</th>
                <th className="px-4 py-2">Success</th>
                <th className="px-4 py-2">Fail</th>
                <th className="px-4 py-2">Latency</th>
                <th className="px-4 py-2">Last Test</th>
                <th className="px-4 py-2">Alive</th>
              </tr>
            </thead>
            <tbody>
              {stats.map((s) => (
                <tr key={s.url} className="border-b border-gray-700">
                  <td className="px-4 py-2">{s.url}</td>
                  <td className="px-4 py-2">{s.score}</td>
                  <td className="px-4 py-2">{s.usage_count}</td>
                  <td className="px-4 py-2 text-green-400">
                    {s.success_count}
                  </td>
                  <td className="px-4 py-2 text-red-400">{s.fail_count}</td>
                  <td className="px-4 py-2">{s.latency_ms} ms</td>
                  <td className="px-4 py-2">{s.last_test}</td>
                  <td className="px-4 py-2">
                    <span
                      className={`px-2 py-1 rounded text-xs ${
                        s.alive ? "bg-green-600" : "bg-red-600"
                      }`}
                    >
                      {s.alive ? "Alive" : "Dead"}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </DashboardLayout>
  );
}
