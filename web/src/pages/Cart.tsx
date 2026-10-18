import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { errorMessage } from "../api";
import { useAuth } from "../auth";
import { useCart } from "../cart";
import { Alert, Loading } from "../components/Feedback";
import { Page } from "../components/Layout";
import { Price } from "../components/Price";
import { ProductImage } from "../components/ProductImage";
import { describeAttributes, formatMoney } from "../format";
import { useTitle } from "../hooks";
import type { CartItem } from "../types";

function Line({ item }: { item: CartItem }) {
  const { update, remove } = useCart();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const max = Math.max(item.quantity, Math.min(item.available, 10));

  async function run(fn: () => Promise<void>) {
    setBusy(true);
    setError(null);
    try {
      await fn();
    } catch (e) {
      setError(errorMessage(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <li className={`flex gap-4 border-b border-line py-4 last:border-0 ${busy ? "opacity-60" : ""}`}>
      <Link to={`/p/${item.slug}`} className="h-28 w-28 shrink-0 overflow-hidden rounded-[3px] sm:h-36 sm:w-36">
        <ProductImage src={item.image_url} category={item.category_slug} alt={item.title} size="sm" />
      </Link>
      <div className="min-w-0 flex-1">
        <div className="flex flex-col justify-between gap-1 sm:flex-row">
          <Link to={`/p/${item.slug}`} className="text-[17px] leading-snug hover:text-amber-dark">{item.title}</Link>
          <div className="shrink-0 sm:text-right"><Price cents={item.unit_price} size="sm" /></div>
        </div>
        {!item.purchasable ? (
          <p className="mt-1 text-sm text-bad">No longer available. Remove it to check out.</p>
        ) : item.available < item.quantity ? (
          <p className="mt-1 text-sm text-bad">Only {item.available} licenses left. Lower the quantity to continue.</p>
        ) : (
          <p className="mt-1 text-xs text-ok">Available</p>
        )}
        <p className="mt-0.5 text-xs text-muted">Sold by {item.store_name}</p>
        {describeAttributes(item.attributes) && <p className="mt-0.5 text-xs"><span className="font-semibold">{describeAttributes(item.attributes)}</span></p>}
        <div className="mt-3 flex flex-wrap items-center gap-3 text-sm">
          <label className="flex items-center gap-2">
            <span className="sr-only">Quantity</span>
            <select
              value={item.quantity}
              disabled={busy || !item.purchasable}
              onChange={(e) => run(() => update(item.variant_id, Number(e.target.value)))}
              className="rounded-[8px] border border-line bg-page px-2 py-1 shadow-sm"
            >
              {Array.from({ length: max }, (_, i) => i + 1).map((n) => (
                <option key={n} value={n}>Qty: {n}</option>
              ))}
            </select>
          </label>
          <span className="h-4 border-l border-line" aria-hidden />
          <button onClick={() => run(() => remove(item.variant_id))} disabled={busy} className="link">Delete</button>
        </div>
        {error && <div className="mt-2"><Alert>{error}</Alert></div>}
      </div>
    </li>
  );
}

export default function CartPage() {
  useTitle("Shopping cart");
  const { user, loading } = useAuth();
  const { cart } = useCart();
  const navigate = useNavigate();

  if (loading) return <Page><Loading /></Page>;
  if (!user) {
    return (
      <Page narrow>
        <div className="panel p-8">
          <h1 className="text-[22px] font-bold">Your cart is empty</h1>
          <p className="mt-2 text-muted">Sign in to see items you saved or start shopping.</p>
          <div className="mt-5 flex gap-3">
            <Link to="/signin?next=/cart" className="btn-primary">Sign in to your account</Link>
            <Link to="/register" className="btn-quiet">Create account</Link>
          </div>
        </div>
      </Page>
    );
  }
  if (!cart) return <Page><Loading /></Page>;

  const blocked = cart.items.some((i) => !i.purchasable || i.available < i.quantity);
  const summary = (
    <div className="panel p-5">
      <p className="text-[18px]">
        Subtotal ({cart.item_count} {cart.item_count === 1 ? "item" : "items"}): <span className="font-bold tnum">{formatMoney(cart.subtotal)}</span>
      </p>
      <button className="btn-primary mt-4 w-full" disabled={cart.items.length === 0 || blocked} onClick={() => navigate("/checkout")}>
        Proceed to checkout
      </button>
      {blocked && <p className="mt-2 text-xs text-bad">Fix the items marked in red before checking out.</p>}
    </div>
  );

  return (
    <Page>
      <div className="grid gap-5 lg:grid-cols-[1fr_300px]">
        <section className="bg-white p-5">
          <h1 className="text-[26px] font-normal">Shopping cart</h1>
          {cart.items.length === 0 ? (
            <div className="py-6">
              <p className="text-muted">Your cart is empty.</p>
              <Link to="/s" className="btn-primary mt-4">Continue shopping</Link>
            </div>
          ) : (
            <>
              <p className="mt-1 border-b border-line pb-2 text-right text-sm text-muted">Price</p>
              <ul>{cart.items.map((it) => <Line key={it.variant_id} item={it} />)}</ul>
              <p className="border-t border-line pt-3 text-right text-[18px]">
                Subtotal ({cart.item_count} {cart.item_count === 1 ? "item" : "items"}): <span className="font-bold tnum">{formatMoney(cart.subtotal)}</span>
              </p>
            </>
          )}
        </section>
        {cart.items.length > 0 && <aside className="lg:sticky lg:top-32 lg:self-start">{summary}</aside>}
      </div>
    </Page>
  );
}
