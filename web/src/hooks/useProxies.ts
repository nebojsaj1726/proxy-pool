import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, type Proxy, type ProxyStats } from "../api/api";
import { useState } from "react";

export function useProxyList() {
  return useQuery<Proxy[], Error>({
    queryKey: ["proxies"],
    queryFn: api.listProxies,
    refetchInterval: 5000,
    staleTime: 5000,
    refetchOnWindowFocus: false,
  });
}

export function useProxyStats() {
  return useQuery<ProxyStats[], Error>({
    queryKey: ["proxyStats"],
    queryFn: api.getStats,
    refetchInterval: 5000,
    staleTime: 5000,
    refetchOnWindowFocus: false,
  });
}

export function useAllocateProxy() {
  const [allocated, setAllocated] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: api.allocateProxy,
    onSuccess: (data) => {
      setAllocated(data.allocated);
      queryClient.invalidateQueries({ queryKey: ["proxies"] });
    },
  });

  const clearAllocated = () => setAllocated(null);

  return {
    allocate: mutation.mutateAsync,
    allocating: mutation.isPending,
    allocateError: mutation.error,
    allocated,
    clearAllocated,
  };
}
