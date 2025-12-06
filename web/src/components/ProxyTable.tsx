import { ProxyRow } from "./ProxyRow";
import type { Proxy } from "../api/api";

interface Props {
  proxies: Proxy[];
}

export function ProxyTable({ proxies }: Props) {
  return (
    <div className="overflow-auto rounded-lg border border-gray-700 bg-darkCard">
      <table className="w-full text-left text-darkText">
        <thead className="bg-gray-800">
          <tr>
            <th className="px-4 py-2">Proxy</th>
            <th className="px-4 py-2">Status</th>
            <th className="px-4 py-2">Last Test</th>
          </tr>
        </thead>
        <tbody>
          {proxies.map((p) => (
            <ProxyRow key={p.url} proxy={p} />
          ))}
        </tbody>
      </table>
    </div>
  );
}
