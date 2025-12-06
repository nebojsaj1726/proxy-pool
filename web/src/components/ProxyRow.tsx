import { type Proxy } from "../api/api";

interface Props {
  proxy: Proxy;
}

export function ProxyRow({ proxy }: Props) {
  return (
    <tr className="border-b border-gray-700">
      <td className="px-4 py-2">{proxy.url}</td>

      <td className="px-4 py-2">
        <span
          className={`px-2 py-1 rounded text-xs ${
            proxy.alive ? "bg-green-600" : "bg-red-600"
          }`}
        >
          {proxy.alive ? "Alive" : "Dead"}
        </span>
      </td>

      <td className="px-4 py-2">{proxy.last_test}</td>
    </tr>
  );
}
