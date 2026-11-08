"use client";

import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import TransactionList from "@/components/TransactionList";
import TransferForm from "@/components/TransferForm";
import {
  api,
  ApiError,
  clearToken,
  getToken,
  UNAUTHORIZED_EVENT,
  type HistoryItem,
} from "@/lib/api";
import { formatRupiah } from "@/lib/format";

export default function DashboardPage() {
  const router = useRouter();
  const [message, setMessage] = useState("");
  const [balance, setBalance] = useState<number | null>(null);
  const [items, setItems] = useState<HistoryItem[]>([]);
  const [loadError, setLoadError] = useState("");

  const load = useCallback(async () => {
    try {
      const [me, wallet, history] = await Promise.all([
        api.me(),
        api.wallet(),
        api.transactions(),
      ]);
      setMessage(me.message);
      setBalance(wallet.balance);
      setItems(history.items);
      setLoadError("");
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) return; // ditangani event 401
      setLoadError(err instanceof ApiError ? err.message : "Unexpected error occurred.");
    }
  }, []);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }

    // Token ditolak server (kedaluwarsa): kembali ke login.
    const onUnauthorized = () => window.location.assign("/login");
    window.addEventListener(UNAUTHORIZED_EVENT, onUnauthorized);

    load();
    return () => window.removeEventListener(UNAUTHORIZED_EVENT, onUnauthorized);
  }, [router, load]);

  function logout() {
    clearToken();
    window.location.assign("/login");
  }

  return (
    <main className="mx-auto max-w-xl space-y-4 bg-zinc-50 p-4 sm:p-6">
      <header className="flex items-start justify-between gap-4">
        <h1 className="text-xl font-semibold text-zinc-900 sm:text-2xl">
          {message || "Loading..."}
        </h1>
        <button
          onClick={logout}
          className="shrink-0 rounded-md border border-zinc-300 bg-white px-3 py-1.5 text-sm text-zinc-700 hover:bg-zinc-100"
        >
          Logout
        </button>
      </header>

      {loadError && (
        <div role="alert" className="flex items-center justify-between gap-3 rounded-md bg-red-50 p-3 text-sm text-red-700">
          <span>{loadError}</span>
          <button onClick={load} className="shrink-0 font-medium underline">
            Try again
          </button>
        </div>
      )}

      <section className="rounded-xl bg-blue-600 p-6 text-white shadow">
        <p className="text-sm opacity-80">Balance</p>
        <p className="mt-1 text-3xl font-semibold">
          {balance === null ? "..." : formatRupiah(balance)}
        </p>
      </section>

      <TransferForm onSuccess={load} />
      <TransactionList items={items} />
    </main>
  );
}