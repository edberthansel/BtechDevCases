"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { api, ApiError, clearToken, getToken } from "@/lib/api";

export default function DashboardPage() {
  const router = useRouter();
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    api
      .me()
      .then((res) => setMessage(res.message))
      .catch((err) => {
        if (err instanceof ApiError && err.status === 401) router.replace("/login");
        else setError(err instanceof ApiError ? err.message : "Unexpected error occured.");
      });
  }, [router]);

  function logout() {
    clearToken();
    window.location.assign("/login");
  }

  return (
    <main className="mx-auto max-w-xl p-6">
      <h1 className="text-2xl font-semibold text-zinc-900">{message || "Loading..."}</h1>
      {error && <p className="mt-2 text-sm text-red-600">{error}</p>}
      <button onClick={logout} className="mt-6 rounded-md border px-4 py-2 text-sm hover:bg-zinc-100">
        Logout
      </button>
    </main>
  );
}