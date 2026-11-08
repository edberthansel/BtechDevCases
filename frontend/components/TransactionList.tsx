import type { HistoryItem } from "@/lib/api";
import { formatDateTime, formatRupiah } from "@/lib/format";

export default function TransactionList({ items }: { items: HistoryItem[] }) {
  return (
    <section className="rounded-xl bg-white p-6 shadow">
      <h2 className="mb-4 text-lg font-semibold text-zinc-900">Transaction history</h2>

      {items.length === 0 ? (
        <p className="text-sm text-zinc-500">No transactions yet.</p>
      ) : (
        <ul className="divide-y divide-zinc-100">
          {items.map((item) => {
            const incoming = item.direction === "in";
            return (
              <li key={item.id} className="flex items-start justify-between gap-4 py-3">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium text-zinc-900">
                    {incoming ? "From " : "To "}
                    {item.counterparty}
                  </p>
                  {item.notes && <p className="truncate text-xs text-zinc-600">{item.notes}</p>}
                  <p className="text-xs text-zinc-400">{formatDateTime(item.createdAt)}</p>
                </div>
                <p
                  className={`shrink-0 text-sm font-semibold ${
                    incoming ? "text-green-600" : "text-red-600"
                  }`}
                >
                  {incoming ? "+" : "-"}
                  {formatRupiah(item.amount)}
                </p>
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}