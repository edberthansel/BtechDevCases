"use client";

import { useEffect, useRef } from "react";
import { api, clearToken } from "@/lib/api";

const IDLE_MINUTES = Number(process.env.NEXT_PUBLIC_IDLE_MINUTES ?? 15);
const IDLE_LIMIT_MS = IDLE_MINUTES * 60 * 1000;
const CHECK_EVERY_MS = 10_000;
const KEEPALIVE_EVERY_MS = 5 * 60 * 1000;

const ACTIVITY_EVENTS = ["mousemove", "mousedown", "keydown", "touchstart", "scroll", "click"] as const;

export function useIdleLogout(onIdle: () => void) {
  const onIdleRef = useRef(onIdle);
  onIdleRef.current = onIdle;

  useEffect(() => {
    let lastActivity = Date.now();
    let lastKeepAlive = Date.now();
    let fired = false;

    const markActive = () => {
      const now = Date.now();
      lastActivity = now;
      // Perpanjang token di server selama user aktif.
      if (now - lastKeepAlive >= KEEPALIVE_EVERY_MS) {
        lastKeepAlive = now;
        api.me().catch(() => {});
      }
    };

    const check = () => {
      if (!fired && Date.now() - lastActivity >= IDLE_LIMIT_MS) {
        fired = true;
        clearToken();
        onIdleRef.current();
      }
    };

    const onVisible = () => {
      if (document.visibilityState === "visible") check();
    };

    ACTIVITY_EVENTS.forEach((ev) => window.addEventListener(ev, markActive, { passive: true }));
    document.addEventListener("visibilitychange", onVisible);
    const interval = setInterval(check, CHECK_EVERY_MS);

    return () => {
      ACTIVITY_EVENTS.forEach((ev) => window.removeEventListener(ev, markActive));
      document.removeEventListener("visibilitychange", onVisible);
      clearInterval(interval);
    };
  }, []);
}