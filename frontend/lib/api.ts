const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
const TOKEN_KEY = "btech_token";
const REQUEST_TIMEOUT_MS = 10_000;

export const UNAUTHORIZED_EVENT = "btech:unauthorized";

// ---------- penyimpanan token ----------

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  try {
    return localStorage.getItem(TOKEN_KEY);
  } catch {
    return null;
  }
}

export function setToken(token: string) {
  try {
    localStorage.setItem(TOKEN_KEY, token);
  } catch {
  }
}

export function clearToken() {
  try {
    localStorage.removeItem(TOKEN_KEY);
  } catch {
  }
}


export class ApiError extends Error {
  status: number; // 0 = error jaringan atau timeout
  fields?: Record<string, string>;

  constructor(status: number, message: string, fields?: Record<string, string>) {
    super(message);
    this.status = status;
    this.fields = fields;
  }
}


type RequestOptions = {
  method?: "GET" | "POST";
  body?: unknown;
  auth?: boolean; // default true: kirim token
  headers?: Record<string, string>;
  retries?: number; // jumlah pengulangan tambahan
};

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

export async function apiRequest<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const { method = "GET", body, auth = true, headers = {}, retries = 0 } = opts;

  for (let attempt = 0; ; attempt++) {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);
    let retryable = false;

    try {
      const h: Record<string, string> = { ...headers };
      if (body !== undefined) h["Content-Type"] = "application/json";
      if (auth) {
        const token = getToken();
        if (token) h["Authorization"] = `Bearer ${token}`;
      }

      const res = await fetch(`${API_URL}${path}`, {
        method,
        headers: h,
        body: body !== undefined ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      });

      // Sliding expiration: simpan token baru kalau server mengirimnya.
      const refreshed = res.headers.get("X-Refreshed-Token");
      if (refreshed && auth) setToken(refreshed);

      const data = await res.json().catch(() => null);

      if (res.ok) return data as T;

      if (res.status === 401 && auth) {
        clearToken();
        window.dispatchEvent(new Event(UNAUTHORIZED_EVENT));
      }

      if (res.status >= 500 && attempt < retries) {
        retryable = true; // error server: coba lagi
      } else {
        throw new ApiError(res.status, data?.error ?? "Terjadi kesalahan", data?.fields);
      }
    } catch (e) {
      if (e instanceof ApiError) throw e;
      // Gagal jaringan atau timeout.
      if (attempt >= retries) {
        throw new ApiError(0, "Koneksi lemah atau terputus. Silakan coba lagi.");
      }
      retryable = true;
    } finally {
      clearTimeout(timer);
    }

    if (retryable) await sleep(Math.min(1000 * 2 ** attempt, 8000));
  }
}


export type MeResponse = { message: string; id: string; email: string };
export type WalletResponse = { balance: number; currency: string };
export type HistoryItem = {
  id: string;
  direction: "in" | "out";
  counterparty: string;
  amount: number;
  notes: string;
  createdAt: string;
};
export type TransferResponse = {
  id: string;
  recipient: string;
  amount: number;
  notes: string;
  createdAt: string;
  replayed: boolean;
};
export type TransferInput = { recipient: string; amount: number; notes: string };

// ---------- fungsi per endpoint ----------

export const api = {
  register: (email: string, password: string, confirmPassword: string) =>
    apiRequest<{ id: string; email: string }>("/register", {
      method: "POST",
      auth: false,
      body: { email, password, confirmPassword },
    }),

  login: (email: string, password: string) =>
    apiRequest<{ token: string; expiresAt: string; user: { id: string; email: string } }>(
      "/login",
      { method: "POST", auth: false, body: { email, password } },
    ),

  me: () => apiRequest<MeResponse>("/me", { retries: 2 }),

  wallet: () => apiRequest<WalletResponse>("/wallet", { retries: 2 }),

  transactions: () =>
    apiRequest<{ items: HistoryItem[] }>("/transactions?limit=20", { retries: 2 }),

  // Aman di-retry karena memakai Idempotency-Key yang SAMA di setiap percobaan.
  transfer: (input: TransferInput, idempotencyKey: string) =>
    apiRequest<TransferResponse>("/transfer", {
      method: "POST",
      body: input,
      headers: { "Idempotency-Key": idempotencyKey },
      retries: 3,
    }),
};