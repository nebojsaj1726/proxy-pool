import { DashboardLayout } from "../layouts/DashboardLayout";
import { ProxyTable } from "../components/ProxyTable";
import { Skeleton } from "../components/Skeleton";
import { Button } from "../components/Button";
import { useAllocateProxy, useProxyList } from "../hooks/useProxies";
import { useEffect } from "react";

export default function ProxiesPage() {
  const { data, isLoading, isError, error } = useProxyList();
  const { allocate, allocated, allocating, allocateError, clearAllocated } =
    useAllocateProxy();

  useEffect(() => {
    if (!allocated || !data) return;

    const proxy = data.find((p) => p.url === allocated);

    if (!proxy || !proxy.alive) {
      clearAllocated();
    }
  }, [allocated, data, clearAllocated]);

  return (
    <DashboardLayout>
      <h1 className="text-2xl font-bold mb-4 text-darkText">Proxies</h1>
      {isLoading && <Skeleton className="h-64 w-full" />}
      {isError && (
        <div className="text-red-400">{error?.message || "Failed to load"}</div>
      )}
      {data && <ProxyTable proxies={data} />}

      <Button
        onClick={() => allocate()}
        disabled={allocating}
        className="mt-8 mb-2 w-40 p-2.5"
      >
        {allocating ? "Allocating..." : "Allocate Proxy"}
      </Button>
      {allocated && (
        <div className="text-green-400 mb-4">Allocated: {allocated}</div>
      )}
      {allocateError && (
        <div className="text-red-400 mb-4">
          {(allocateError as Error).message}
        </div>
      )}
    </DashboardLayout>
  );
}
