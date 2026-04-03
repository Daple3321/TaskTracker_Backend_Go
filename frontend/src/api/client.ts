const getApiUrl = (): string => {
  const base = import.meta.env.VITE_API_URL;
  if (typeof base === "string" && base.length > 0) return base.replace(/\/$/, "");
  return "http://localhost:8080";
};

export type ApiRequestInit = RequestInit & { token?: string | null };

export async function apiFetch(
  path: string,
  init: ApiRequestInit = {}
): Promise<Response> {
  const { token, ...rest } = init;
  const url = `${getApiUrl()}${path.startsWith("/") ? path : `/${path}`}`;
  const headers = new Headers(rest.headers);
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  if (rest.body && typeof rest.body === "string" && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  return fetch(url, { ...rest, headers });
}

export { getApiUrl };
