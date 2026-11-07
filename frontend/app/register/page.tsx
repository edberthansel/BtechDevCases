"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState, type FormEvent } from "react";
import Field from "@/components/Field";
import { api, ApiError, getToken, setToken } from "@/lib/api";

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export default function RegisterPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (getToken()) router.replace("/dashboard");
  }, [router]);

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (loading) return;

    // Validasi cepat di browser (server tetap memvalidasi ulang).
    const newErrors: Record<string, string> = {};
    if (!EMAIL_PATTERN.test(email.trim())) newErrors.email = "Email format is not valid!";
    if (password.length < 8) newErrors.password = "Password should have minimum 8 character!";
    if (confirmPassword !== password) newErrors.confirmPassword = "Password does not match!";
    setErrors(newErrors);
    setFormError("");
    if (Object.keys(newErrors).length > 0) return;

    setLoading(true);
    try {
      await api.register(email, password, confirmPassword);
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 409) setErrors({ email: "Email is already registered!" });
        else if (err.fields) setErrors(err.fields);
        else setFormError(err.message);
      } else {
        setFormError("An unexpected error occurred. Please contact the administrator!");
      }
      setLoading(false);
      return;
    }

    // Register sukses: login otomatis.
    try {
      const res = await api.login(email, password);
      setToken(res.token);
      router.replace("/dashboard");
    } catch {
      router.replace("/login"); 
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-zinc-50 p-4">
      <form
        onSubmit={onSubmit}
        noValidate
        className="w-full max-w-sm space-y-4 rounded-xl bg-white p-6 shadow"
      >
        <h1 className="text-xl font-semibold text-zinc-900">Register</h1>

        {formError && (
          <div role="alert" className="rounded-md bg-red-50 p-3 text-sm text-red-700">
            {formError}
          </div>
        )}

        <Field
          label="Email"
          name="email"
          type="email"
          autoComplete="email"
          value={email}
          onChange={setEmail}
          error={errors.email}
          disabled={loading}
        />
        <Field
          label="Password"
          name="password"
          type="password"
          autoComplete="new-password"
          value={password}
          onChange={setPassword}
          error={errors.password}
          disabled={loading}
        />
        <Field
          label="Confirm Password"
          name="confirmPassword"
          type="password"
          autoComplete="new-password"
          value={confirmPassword}
          onChange={setConfirmPassword}
          error={errors.confirmPassword}
          disabled={loading}
        />

        <button
          type="submit"
          disabled={loading}
          className="w-full rounded-md bg-blue-600 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60"
        >
          {loading ? "Loading..." : "Register"}
        </button>

        <p className="text-center text-sm text-zinc-600">
          Already have account?{" "}
          <Link href="/login" className="text-blue-600 hover:underline">
            Login
          </Link>
        </p>
      </form>
    </main>
  );
}