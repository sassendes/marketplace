import { Link, useParams, useSearchParams } from "react-router-dom";
import { Alert, Loading } from "../components/Feedback";
import { Page } from "../components/Layout";
import { ProductImage } from "../components/ProductImage";
import { formatDate, formatMoney } from "../format";
import { useFetch, useTitle } from "../hooks";
import type { Order } from "../types";

const statusText: Record<string, { label: string; tone: string }> = {
  paid: { label: "Paid, awaiting delivery from the seller", tone: "text-ok" },
  pending: { label: "Processing", tone: "text-amber-dark" },
  cancelled: { label: "Cancelled, not charged", tone: "text-bad" },
};

function StatusLine({ status }: { status: string }) {
  const s = statusText[status] ?? { label: status, tone: "" };
  return <span className={`font-bold ${s.tone}`}>{s.label}</span>;
}

export function OrdersPage() {
  useTitle("Your orders");
  const { data, loading, error } = useFetch<Order[]>("/orders");
  return (
    <Page narrow>
      <h1 className="text-[26px] font-normal">Your orders</h1>
      {loading && <Loading />}
      {error && <Alert>{error}</Alert>}
      {data && data.length === 0 && (
        <div className="panel mt-4 p-6">
          <p>You have not placed any orders yet.</p>
          <Link to="/s" className="btn-primary mt-4">Start shopping</Link>
        </div>
      )}
      <ul className="mt-4 space-y-4">
        {data?.map((o) => (
          <li key={o.id} className="overflow-hidden rounded-[8px] border border-line bg-white">
            <div className="grid grid-cols-2 gap-3 bg-page px-5 py-3 text-xs text-muted sm:grid-cols-4">
              <div><p>Order placed</p><p className="text-sm text-ink">{formatDate(o.created_at)}</p></div>
              <div><p>Total</p><p className="tnum text-sm text-ink">{formatMoney(o.total_cents)}</p></div>
              <div><p>Items</p><p className="tnum text-sm text-ink">{o.item_count}</p></div>
              <div className="sm:text-right"><p className="tnum">Order #{o.id}</p><Link to={`/orders/${o.id}`} className="link text-sm">View order details</Link></div>
            </div>
            <div className="px-5 py-4">
              <StatusLine status={o.status} />
              <ul className="mt-3 flex flex-wrap gap-4">
                {o.preview?.map((p, i) => (
                  <li key={i} className="flex w-full items-center gap-3 sm:w-[48%]">
                    <div className="h-16 w-16 shrink-0 overflow-hidden rounded-[3px]"><ProductImage src={p.image_url} category={p.category_slug} alt={p.title} size="sm" /></div>
                    {p.slug ? <Link to={`/p/${p.slug}`} className="link line-clamp-2 text-sm">{p.title}</Link> : <span className="text-sm">{p.title}</span>}
                  </li>
                ))}
              </ul>
            </div>
          </li>
        ))}
      </ul>
    </Page>
  );
}

export function OrderDetailPage() {
  const { id } = useParams();
  const [params] = useSearchParams();
  const { data: o, loading, error } = useFetch<Order>(`/orders/${id}`);
  useTitle(o ? `Order #${o.id}` : "Order");
  if (loading) return <Page><Loading /></Page>;
  if (error || !o) return <Page narrow><Alert>{error ?? "Order not found"}</Alert></Page>;
  return (
    <Page narrow>
      {params.get("placed") && o.status === "paid" && (
        <div className="mb-5">
          <Alert tone="ok">
            <span className="font-bold">Order placed, thanks.</span> Each seller will deliver your software to your account email.
          </Alert>
        </div>
      )}
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <h1 className="text-[26px] font-normal">Order details</h1>
        <Link to="/orders" className="link text-sm">Back to your orders</Link>
      </div>
      <p className="mt-1 text-sm text-muted tnum">Ordered on {formatDate(o.created_at)}, order #{o.id}</p>
      <div className="panel mt-4 grid gap-4 p-5 sm:grid-cols-2">
        <div>
          <h2 className="text-sm font-bold">Status</h2>
          <p className="mt-1"><StatusLine status={o.status} /></p>
        </div>
        <div className="sm:text-right">
          <h2 className="text-sm font-bold">Order total</h2>
          <p className="mt-1 text-[18px] font-bold tnum">{formatMoney(o.total_cents)}</p>
        </div>
      </div>
      <div className="mt-4 space-y-4">
        {o.sub_orders?.map((so) => (
          <section key={so.id} className="panel">
            <div className="flex justify-between border-b border-line bg-page px-5 py-2 text-sm">
              <span>Sold by <span className="font-semibold">{so.store_name}</span></span>
              <span className="tnum">{formatMoney(so.subtotal_cents)}</span>
            </div>
            <ul className="divide-y divide-line">
              {so.items?.map((it) => (
                <li key={it.id} className="flex gap-4 px-5 py-4">
                  <div className="h-20 w-20 shrink-0 overflow-hidden rounded-[3px]"><ProductImage src={it.image_url} category={it.category_slug} alt={it.title} size="sm" /></div>
                  <div className="min-w-0 flex-1 text-sm">
                    {it.slug ? <Link to={`/p/${it.slug}`} className="link font-medium">{it.title}</Link> : <span>{it.title}</span>}
                    <p className="mt-1 text-muted tnum">Qty {it.quantity} at {formatMoney(it.unit_price_cents)}</p>
                  </div>
                  <p className="tnum text-sm font-semibold">{formatMoney(it.unit_price_cents * it.quantity)}</p>
                </li>
              ))}
            </ul>
          </section>
        ))}
      </div>
    </Page>
  );
}
