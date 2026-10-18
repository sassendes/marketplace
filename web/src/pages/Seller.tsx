import { useEffect, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { Plus, Trash2 } from "lucide-react";
import { ApiError, api, errorMessage } from "../api";
import { useAuth } from "../auth";
import { Alert, Loading } from "../components/Feedback";
import { Page } from "../components/Layout";
import { ProductImage } from "../components/ProductImage";
import { formatDate, formatMoney } from "../format";
import { useFetch, useTitle } from "../hooks";
import type { Category, Product, ProductList, Seller, SubOrder } from "../types";

function Onboard({ onDone }: { onDone: () => void }) {
  const { reload } = useAuth();
  const [store, setStore] = useState("");
  const [description, setDescription] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await api("/sellers/onboard", { method: "POST", json: { store_name: store, description } });
      await reload();
      onDone();
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={submit} className="panel max-w-xl space-y-4 p-6">
      <div>
        <h2 className="text-[19px] font-bold">Open your store</h2>
        <p className="mt-1 text-sm text-muted">Pick a name buyers will see on every listing. You deliver keys or downloads yourself after each sale. Read the seller guide in the store first.</p>
      </div>
      <div>
        <label htmlFor="store" className="label">Store name</label>
        <input id="store" className="field" value={store} onChange={(e) => setStore(e.target.value)} maxLength={80} />
      </div>
      <div>
        <label htmlFor="sdesc" className="label">About your store</label>
        <textarea id="sdesc" className="field min-h-24" value={description} onChange={(e) => setDescription(e.target.value)} maxLength={2000} />
      </div>
      {error && <Alert>{error}</Alert>}
      <button className="btn-primary" disabled={busy}>{busy ? "Creating store" : "Create store"}</button>
    </form>
  );
}

function ConnectPayouts({ onDone }: { onDone: () => void }) {
  const [acct, setAcct] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await api("/sellers/connect-stripe", { method: "POST", json: { stripe_account_id: acct } });
      onDone();
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={submit} className="panel max-w-xl space-y-4 p-6">
      <div>
        <h2 className="text-[19px] font-bold">Connect payouts</h2>
        <p className="mt-1 text-sm text-muted">Enter your Stripe Connect account ID. In test mode any ID in the format acct_ followed by 8 or more letters or digits works.</p>
      </div>
      <div>
        <label htmlFor="acct" className="label">Stripe account ID</label>
        <input id="acct" className="field" value={acct} onChange={(e) => setAcct(e.target.value)} placeholder="acct_1AbCdEfGh" />
      </div>
      {error && <Alert>{error}</Alert>}
      <div className="flex flex-wrap gap-3">
        <button className="btn-primary" disabled={busy || !acct}>{busy ? "Connecting" : "Connect payouts"}</button>
        <button type="button" className="btn-quiet" onClick={() => setAcct(`acct_test${Math.random().toString(36).slice(2, 12)}`)}>
          Fill a test account ID
        </button>
      </div>
    </form>
  );
}

interface VariantDraft {
  option: string;
  price: string;
  stock: string;
}

function NewProduct({ onCreated }: { onCreated: () => void }) {
  const { data: categories } = useFetch<Category[]>("/catalog/categories");
  const [title, setTitle] = useState("");
  const [brand, setBrand] = useState("");
  const [category, setCategory] = useState("");
  const [description, setDescription] = useState("");
  const [image, setImage] = useState("");
  const [variants, setVariants] = useState<VariantDraft[]>([{ option: "", price: "", stock: "10" }]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function setVariant(i: number, patch: Partial<VariantDraft>) {
    setVariants((vs) => vs.map((v, j) => (j === i ? { ...v, ...patch } : v)));
  }

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    const parsed = variants.map((v, i) => ({
      sku: v.option ? `${v.option.toUpperCase().replace(/[^A-Z0-9]+/g, "-")}-${i + 1}` : `SKU-${i + 1}`,
      price: Math.round(Number(v.price) * 100),
      stock: Number(v.stock),
      attributes: v.option ? { option: v.option } : {},
    }));
    if (parsed.some((v) => !Number.isFinite(v.price) || v.price <= 0)) {
      setError("Every option needs a price above zero.");
      return;
    }
    if (parsed.some((v) => !Number.isInteger(v.stock) || v.stock < 0)) {
      setError("Stock must be a whole number of zero or more.");
      return;
    }
    setBusy(true);
    try {
      await api("/catalog/products", {
        method: "POST",
        json: { title, brand, description, image_url: image, category_id: category ? Number(category) : null, variants: parsed },
      });
      setTitle("");
      setBrand("");
      setDescription("");
      setImage("");
      setVariants([{ option: "", price: "", stock: "10" }]);
      onCreated();
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={submit} className="panel space-y-4 p-6">
      <h2 className="text-[19px] font-bold">Add a product</h2>
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="sm:col-span-2">
          <label htmlFor="ptitle" className="label">Title</label>
          <input id="ptitle" className="field" value={title} onChange={(e) => setTitle(e.target.value)} maxLength={200} />
        </div>
        <div>
          <label htmlFor="pbrand" className="label">Brand</label>
          <input id="pbrand" className="field" value={brand} onChange={(e) => setBrand(e.target.value)} maxLength={100} />
        </div>
        <div>
          <label htmlFor="pcat" className="label">Category</label>
          <select id="pcat" className="field" value={category} onChange={(e) => setCategory(e.target.value)}>
            <option value="">Choose a category</option>
            {categories?.map((c) => (
              <option key={c.id} value={c.id}>{c.parent_id ? `  ${c.name}` : c.name}</option>
            ))}
          </select>
        </div>
        <div className="sm:col-span-2">
          <label htmlFor="pdesc" className="label">Description</label>
          <textarea id="pdesc" className="field min-h-24" value={description} onChange={(e) => setDescription(e.target.value)} maxLength={10000} />
        </div>
        <div className="sm:col-span-2">
          <label htmlFor="pimg" className="label">Image URL</label>
          <input id="pimg" className="field" value={image} onChange={(e) => setImage(e.target.value)} placeholder="https://" />
          <p className="mt-1 text-xs text-muted">Optional. Use an https link to a square photo.</p>
        </div>
      </div>
      <fieldset>
        <legend className="label">Editions, price and licenses available</legend>
        <div className="space-y-2">
          {variants.map((v, i) => (
            <div key={i} className="grid grid-cols-[1fr_110px_90px_36px] items-center gap-2">
              <input aria-label={`Option ${i + 1} name`} className="field" placeholder="e.g. Personal, Team, Lifetime" value={v.option} onChange={(e) => setVariant(i, { option: e.target.value })} />
              <input aria-label={`Option ${i + 1} price in dollars`} className="field tnum" inputMode="decimal" placeholder="Price $" value={v.price} onChange={(e) => setVariant(i, { price: e.target.value })} />
              <input aria-label={`Option ${i + 1} stock`} className="field tnum" inputMode="numeric" placeholder="Licenses" value={v.stock} onChange={(e) => setVariant(i, { stock: e.target.value })} />
              <button
                type="button"
                aria-label={`Remove option ${i + 1}`}
                disabled={variants.length === 1}
                onClick={() => setVariants((vs) => vs.filter((_, j) => j !== i))}
                className="flex h-9 w-9 items-center justify-center rounded text-muted hover:bg-page disabled:opacity-30"
              >
                <Trash2 size={16} />
              </button>
            </div>
          ))}
        </div>
        <button type="button" onClick={() => setVariants((vs) => [...vs, { option: "", price: "", stock: "10" }])} className="link mt-2 inline-flex items-center gap-1 text-sm">
          <Plus size={15} aria-hidden /> Add another option
        </button>
      </fieldset>
      {error && <Alert>{error}</Alert>}
      <button className="btn-primary" disabled={busy || !title}>{busy ? "Publishing" : "Publish listing"}</button>
    </form>
  );
}

function StockEditor({ variantId, available, onSaved }: { variantId: number; available: number; onSaved: () => void }) {
  const [value, setValue] = useState(String(available));
  const [state, setState] = useState<"idle" | "saving" | "saved" | "error">("idle");
  useEffect(() => setValue(String(available)), [available]);
  async function save() {
    setState("saving");
    try {
      await api("/stock", { method: "PUT", json: { variant_id: variantId, on_hand: Number(value) } });
      setState("saved");
      onSaved();
    } catch {
      setState("error");
    }
  }
  return (
    <span className="inline-flex items-center gap-2">
      <input aria-label="Licenses available" className="field w-20 py-1 tnum" inputMode="numeric" value={value} onChange={(e) => { setValue(e.target.value); setState("idle"); }} />
      <button onClick={save} disabled={state === "saving" || value === String(available)} className="btn-quiet px-3 py-1 text-xs">Save</button>
      {state === "saved" && <span className="text-xs text-ok">Saved</span>}
      {state === "error" && <span className="text-xs text-bad">Not saved</span>}
    </span>
  );
}

function Listings({ list, reload }: { list: ProductList | null; reload: () => void }) {
  const [error, setError] = useState<string | null>(null);
  async function toggle(p: Product) {
    setError(null);
    try {
      await api(`/catalog/products/${p.id}/status`, { method: "PATCH", json: { status: p.status === "active" ? "inactive" : "active" } });
      reload();
    } catch (e) {
      setError(errorMessage(e));
    }
  }
  if (!list) return <Loading />;
  return (
    <section className="panel">
      <h2 className="border-b border-line px-5 py-3 text-[17px] font-bold">Your listings ({list.total})</h2>
      {error && <div className="p-4"><Alert>{error}</Alert></div>}
      {list.products.length === 0 && <p className="p-5 text-sm text-muted">No listings yet. Add your first product above.</p>}
      <ul className="divide-y divide-line">
        {list.products.map((p) => (
          <li key={p.id} className="flex flex-col gap-3 px-5 py-4 sm:flex-row">
            <div className="h-16 w-16 shrink-0 overflow-hidden rounded-[3px]"><ProductImage src={p.image_url} category={p.category_slug} alt={p.title} size="sm" /></div>
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-2">
                {p.status === "active" ? <Link to={`/p/${p.slug}`} className="link font-medium">{p.title}</Link> : <span className="font-medium">{p.title}</span>}
                <span className={`rounded-[3px] px-1.5 py-0.5 text-xs ${p.status === "active" ? "bg-[#e7f4ec] text-ok" : "bg-page text-muted"}`}>
                  {p.status === "active" ? "Live" : p.status === "removed" ? "Removed by admin" : "Hidden"}
                </span>
              </div>
              <ul className="mt-2 space-y-1.5 text-sm">
                {p.variants?.map((v) => (
                  <li key={v.id} className="flex flex-wrap items-center gap-3">
                    <span className="w-40 truncate text-muted">{Object.values(v.attributes ?? {}).join(" / ") || v.sku}</span>
                    <span className="w-20 tnum">{formatMoney(v.price)}</span>
                    <StockEditor variantId={v.id} available={v.available} onSaved={reload} />
                  </li>
                ))}
              </ul>
            </div>
            {p.status !== "removed" && (
              <button onClick={() => toggle(p)} className="btn-quiet self-start text-xs">{p.status === "active" ? "Hide listing" : "Make live"}</button>
            )}
          </li>
        ))}
      </ul>
    </section>
  );
}

function Sales() {
  const { data } = useFetch<SubOrder[]>("/orders/sales");
  if (!data) return <Loading />;
  const paid = data.filter((s) => s.status === "paid");
  const gross = paid.reduce((n, s) => n + s.subtotal_cents, 0);
  const fees = paid.reduce((n, s) => n + s.commission_cents, 0);
  return (
    <section className="panel">
      <div className="grid grid-cols-3 divide-x divide-line border-b border-line text-center">
        <div className="p-4"><p className="text-xs text-muted">Orders</p><p className="text-[19px] font-bold tnum">{paid.length}</p></div>
        <div className="p-4"><p className="text-xs text-muted">Sales</p><p className="text-[19px] font-bold tnum">{formatMoney(gross)}</p></div>
        <div className="p-4"><p className="text-xs text-muted">Your earnings</p><p className="text-[19px] font-bold tnum">{formatMoney(gross - fees)}</p></div>
      </div>
      {data.length === 0 ? (
        <p className="p-5 text-sm text-muted">No sales yet.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="bg-page text-left text-xs text-muted">
              <tr><th className="px-5 py-2 font-medium">Order</th><th className="px-5 py-2 font-medium">Deliver to</th><th className="px-5 py-2 font-medium">Items</th><th className="px-5 py-2 font-medium">Date</th><th className="px-5 py-2 font-medium">Status</th><th className="px-5 py-2 text-right font-medium">Subtotal</th><th className="px-5 py-2 text-right font-medium">Fee</th></tr>
            </thead>
            <tbody className="divide-y divide-line tnum">
              {data.map((s) => (
                <tr key={s.id}>
                  <td className="px-5 py-2">#{s.order_id}</td>
                  <td className="px-5 py-2"><a className="link" href={`mailto:${s.buyer_email}`}>{s.buyer_email}</a></td>
                  <td className="px-5 py-2">{s.summary}</td>
                  <td className="px-5 py-2">{s.created_at ? formatDate(s.created_at) : ""}</td>
                  <td className="px-5 py-2 capitalize">{s.status}</td>
                  <td className="px-5 py-2 text-right">{formatMoney(s.subtotal_cents)}</td>
                  <td className="px-5 py-2 text-right text-muted">{formatMoney(s.commission_cents)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

export default function SellerDashboard() {
  useTitle("Seller dashboard");
  const { user } = useAuth();
  const [seller, setSeller] = useState<Seller | null | undefined>(undefined);
  const [tab, setTab] = useState<"listings" | "sales">("listings");
  const isSeller = user?.role === "seller";
  const { data: list, reload } = useFetch<ProductList>(isSeller && seller?.status === "active" ? "/catalog/mine" : null, [seller?.status]);

  async function load() {
    try {
      setSeller(await api<Seller>("/sellers/me"));
    } catch (e) {
      if (e instanceof ApiError && e.status === 404) setSeller(null);
      else setSeller(null);
    }
  }
  useEffect(() => {
    load();
  }, [user?.id, user?.role]);

  if (seller === undefined) return <Page><Loading /></Page>;
  return (
    <Page narrow>
      <div className="mb-5 flex flex-wrap items-baseline justify-between gap-2">
        <h1 className="text-[26px] font-normal">{seller ? seller.store_name : "Sell on sassendes.shop"}</h1>
        {seller && (
          <span className={`text-sm font-semibold ${seller.status === "active" ? "text-ok" : seller.status === "suspended" ? "text-bad" : "text-amber-dark"}`}>
            {seller.status === "active" ? "Store is live" : seller.status === "suspended" ? "Store suspended" : "Waiting for payouts"}
          </span>
        )}
      </div>
      {!seller && <Onboard onDone={load} />}
      {seller?.status === "pending" && <ConnectPayouts onDone={load} />}
      {seller?.status === "suspended" && <Alert>Your store was suspended by the marketplace team. Your listings are hidden from buyers.</Alert>}
      {seller?.status === "active" && (
        <div className="space-y-5">
          <div role="tablist" className="flex gap-1 border-b border-line">
            {(["listings", "sales"] as const).map((t) => (
              <button
                key={t}
                role="tab"
                aria-selected={tab === t}
                onClick={() => setTab(t)}
                className={`-mb-px border-b-2 px-4 py-2 text-sm font-semibold ${tab === t ? "border-amber text-ink" : "border-transparent text-muted hover:text-ink"}`}
              >
                {t === "listings" ? "Listings" : "Sales"}
              </button>
            ))}
          </div>
          {tab === "listings" ? (
            <>
              <NewProduct onCreated={reload} />
              <Listings list={list} reload={reload} />
            </>
          ) : (
            <Sales />
          )}
        </div>
      )}
    </Page>
  );
}
