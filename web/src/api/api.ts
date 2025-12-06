import { apiFetch } from "./client";

export interface Proxy {
  url: string;
  alive: boolean;
  last_test: string;
}

export interface ProxyStats extends Proxy {
  score: number;
  usage_count: number;
  fail_count: number;
  success_count: number;
  latency_ms: number;
}

export const api = {
  login: (username: string, password: string) =>
    apiFetch<{ token: string }>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    }),
  register: (username: string, password: string) =>
    apiFetch<void>("/auth/register", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    }),
  listProxies: () => apiFetch<Proxy[]>("/proxies"),
  allocateProxy: () =>
    apiFetch<{ allocated: string }>("/allocate", {
      method: "POST",
    }),
  getStats: () => apiFetch<ProxyStats[]>("/proxies/stats"),
};
