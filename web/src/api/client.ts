export const API_BASE = "http://localhost:8080";

export async function apiFetch<T>(
  url: string,
  options: RequestInit = {}
): Promise<T> {
  const token = localStorage.getItem("token");

  const res = await fetch(API_BASE + url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  });

  if (!res.ok) {
    const msg = await res.text();
    throw new Error(msg || "API error");
  }

  if (res.status === 204) return null as T;
  return res.json();
}
