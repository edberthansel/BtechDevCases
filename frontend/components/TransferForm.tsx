"use client";

import { useRef, useState, type FormEvent } from "react";
import Field from "@/components/Field";
import { api, ApiError } from "@/lib/api";
import { formatRupiah } from "@/lib/format";

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const MAX_AMOUNT = 1_000_000_000;
const MAX_NOTES = 200;

function newIdempotencyKey(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  // Cadangan untuk konteks non-HTTPS. Tetap sesuai aturan server (8-64 karakter).
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 12)}`;
}

export default function TransferForm({ onSuccess }: { onSuccess: () => void }) {
  const [recipient, setRecipient] = useState("");
  const [amount, setAmount] = useState("");
  const [notes, setNotes] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);

  // Key satu percobaan transfer. Disimpan di ref agar tidak memicu render ulang.
  const keyRef = useRef<string | null>(null);

  // Setiap perubahan isi form = transfer yang berbeda = key baru.
  function edit(setter: (v: string) => void) {
    return (value: string) => {
      keyRef.current = null;
      setter(value);
    };
  }

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (loading) return;

    const email = recipient.trim().toLowerCase();
    const value = Number(amount);

    const newErrors: Record<string, string> = {};
    if (!email) newErrors.recipient = "Recipient email is required!";
    else if (!EMAIL_PATTERN.test(email)) newErrors.recipient = "Email format is not valid!";

    if (!amount || !Number.isSafeInteger(value) || value <= 0) {
      newErrors.amount = "Amount must be greater than 0!";
    } else if (value > MAX_AMOUNT) {
      newErrors.amount = `Maximum ${formatRupiah(MAX_AMOUNT)} per transfer!`;
    }

    if (notes.length > MAX_NOTES) newErrors.notes = `Notes can be at most ${MAX_NOTES} characters!`;

    setErrors(newErrors);
    setFormError("");
    setSuccess("");
    if (Object.keys(newErrors).length > 0) return;

    if (!keyRef.current) keyRef.current = newIdempotencyKey();

    setLoading(true);
    try {
      const res = await api.transfer(
        { recipient: email, amount: value, notes: notes.trim() },
        keyRef.current,
      );

      // Sukses: percobaan selesai, key dibuang.
      keyRef.current = null;
      setRecipient("");
      setAmount("");
      setNotes("");
      setSuccess(
        res.replayed
          ? `This transfer was already processed: ${formatRupiah(res.amount)} to ${res.recipient}.`
          : `Transfer successful: ${formatRupiah(res.amount)} to ${res.recipient}.`,
      );
      onSuccess();
    } catch (err) {
      if (err instanceof ApiError && (err.status === 0 || err.status >= 500)) {
        // Hasil tidak pasti: PERTAHANKAN key supaya retry aman.
        setFormError(
          "Weak connection. Your transfer may not have gone through. Press Transfer again to retry. It will not be sent twice.",
        );
      } else {
        // Ditolak jelas oleh server: percobaan ini selesai.
        keyRef.current = null;
        if (err instanceof ApiError) {
          if (err.status === 404) setErrors({ recipient: "Recipient not found!" });
          else if (err.status === 422) setErrors({ amount: "Insufficient balance!" });
          else if (err.status === 401) setFormError("Session expired. Please log in again.");
          else if (err.fields) setErrors(err.fields);
          else setFormError(err.message);
        } else {
          setFormError("An unexpected error occurred. Please try again.");
        }
      }
    } finally {
      setLoading(false);
    }
  }

  const preview = amount && Number(amount) > 0 ? formatRupiah(Number(amount)) : "";

  return (
    <form onSubmit={onSubmit} noValidate className="space-y-4 rounded-xl bg-white p-6 shadow">
      <h2 className="text-lg font-semibold text-zinc-900">Transfer</h2>

      {success && (
        <div role="status" className="rounded-md bg-green-50 p-3 text-sm text-green-700">
          {success}
        </div>
      )}
      {formError && (
        <div role="alert" className="rounded-md bg-red-50 p-3 text-sm text-red-700">
          {formError}
        </div>
      )}

      <Field
        label="Recipient (email)"
        name="recipient"
        type="email"
        autoComplete="off"
        value={recipient}
        onChange={edit(setRecipient)}
        error={errors.recipient}
        disabled={loading}
      />

      <div>
        <Field
          label="Amount (IDR)"
          name="amount"
          autoComplete="off"
          value={amount}
          onChange={edit((v) => setAmount(v.replace(/\D/g, "")))}
          error={errors.amount}
          disabled={loading}
        />
        {preview && !errors.amount && <p className="mt-1 text-xs text-zinc-500">{preview}</p>}
      </div>

      <Field
        label="Notes (optional)"
        name="notes"
        autoComplete="off"
        value={notes}
        onChange={edit(setNotes)}
        error={errors.notes}
        disabled={loading}
      />

      <button
        type="submit"
        disabled={loading}
        className="w-full rounded-md bg-blue-600 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60"
      >
        {loading ? "Sending..." : "Transfer"}
      </button>
    </form>
  );
}