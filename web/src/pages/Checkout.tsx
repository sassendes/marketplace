import { useMemo, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Lock } from "lucide-react";
import { ApiError, api, errorMessage } from "../api";
import { useCart } from "../cart";
import { Alert, Loading } from "../components/Feedback";
import { Page } from "../components/Layout";
import { ProductImage } from "../components/ProductImage";
import { formatMoney } from "../format";
import { useTitle } from "../hooks";
import type { CartItem, CheckoutResult } from "../types";

function newKey() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) return crypto.randomUUID();
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

export default function Checkout() {
  useTitle("Checkout");
  const { cart, refresh, reset } = useCart();
  const navigate = useNavigate();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const key = useRef(newKey());

  const groups = useMemo(() => {
    const m = new Map<string, CartItem[]>();
    for (const it of cart?.items ?? []) m.set(it.store_name, [...(m.get(it.store_name) ?? []), it]);
    return [...m.entries()];
  }, [cart]);

  if (!cart) return <Page><Loading /></Page>;
  if (cart.items.length === 0 && !busy) {
    return (
      <Page narrow>
        <div className="panel p-8">
          <h1 className="text-[22px] font-bold">There is nothing to check out</h1>
          <p className="mt-2 text-muted">Add something to your cart first.</p>
          <Link to="/s" className="btn-primary mt-4">Browse products</Link>
        </div>
      </Page>
    );
  }

  async function place() {
    setBusy(true);
    setError(null);
    try {
      const res = await api<CheckoutResult>("/checkout", { method: "POST", headers: { "Idempotency-Key": key.current } });
      if (res.status === "completed") {
        reset();
        navigate(`/orders/${res.order_id}?placed=1`, { replace: true });
        return;
      }
      setError(res.error ? res.error.charAt(0).toUpperCase() + res.error.slice(1) : "Your order could not be completed.");
      key.current = newKey();
      await refresh();
    } catch (e) {
      if (e instanceof ApiError && e.body && typeof e.body === "object" && "status" in e.body) {
        const r = e.body as CheckoutResult;
        setError(r.error ? `${r.error.charAt(0).toUpperCase()}${r.error.slice(1)}. Nothing was charged.` : "Your order could not be completed.");
      } else {
        setError(errorMessage(e));
      }
      key.current = newKey();
      await refresh().catch(() => undefined);
    } finally {
      setBusy(false);
    }
  }

  return (
    <Page>
      <div className="mb-5 flex items-center justify-between border-b border-line pb-3">
        <h1 className="text-[26px] font-normal">Checkout</h1>
        <span className="flex items-center gap-1 text-sm text-muted"><Lock size={15} aria-hidden /> Secure checkout</span>
      </div>
      <div className="grid gap-5 lg:grid-cols-[1fr_320px]">
        <div className="space-y-5">
          <section className="panel p-5">
            <h2 className="text-[18px] font-bold">Payment method</h2>
            <p className="mt-2 text-sm">Test card ending in 4242</p>
            <p className="mt-1 text-xs text-muted">This store runs in payment test mode. No real money moves.</p>
          </section>
          <section className="panel p-5">
            <h2 className="text-[18px] font-bold">Review items</h2>
            <p className="mt-1 text-sm text-muted">
              You pay once. Each seller delivers their software to your account email after payment.
            </p>
            {groups.map(([store, items]) => (
              <div key={store} className="mt-4 rounded-[var(--radius-panel)] border border-line">
                <p className="border-b border-line bg-page px-4 py-2 text-sm font-semibold">Delivered by {store}</p>
                <ul className="divide-y divide-line">
                  {items.map((it) => (
                    <li key={it.variant_id} className="flex gap-3 px-4 py-3">
                      <div className="h-16 w-16 shrink-0 overflow-hidden rounded-[3px]"><ProductImage src={it.image_url} category={it.category_slug} alt={it.title} size="sm" /></div>
                      <div className="min-w-0 flex-1 text-sm">
                        <p className="line-clamp-2 font-medium">{it.title}</p>
                        <p className="text-muted">Qty {it.quantity}</p>
                      </div>
                      <p className="tnum text-sm font-semibold">{formatMoney(it.unit_price * it.quantity)}</p>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </section>
        </div>
        <aside className="lg:sticky lg:top-32 lg:self-start">
          <div className="panel p-5">
            <button onClick={place} disabled={busy} className="btn-primary w-full">{busy ? "Placing your order" : "Place your order"}</button>
            <p className="mt-2 text-center text-xs text-muted">Licenses are held for you while payment is processed.</p>
            <div className="my-4 border-t border-line" />
            <h2 className="font-bold">Order summary</h2>
            <dl className="mt-2 space-y-1 text-sm tnum">
              <div className="flex justify-between"><dt>Items ({cart.item_count})</dt><dd>{formatMoney(cart.subtotal)}</dd></div>
            </dl>
            <div className="my-3 border-t border-line" />
            <div className="flex justify-between text-[18px] font-bold text-bad tnum"><span>Order total</span><span>{formatMoney(cart.subtotal)}</span></div>
            {error && <div className="mt-4"><Alert>{error}</Alert></div>}
          </div>
        </aside>
      </div>
    </Page>
  );
}
