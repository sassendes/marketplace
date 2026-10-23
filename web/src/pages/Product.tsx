import { useEffect, useMemo, useState, type FormEvent } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { Check, ShieldCheck } from "lucide-react";
import { api, errorMessage } from "../api";
import { useAuth } from "../auth";
import { useCart } from "../cart";
import { Alert, Loading } from "../components/Feedback";
import { Page } from "../components/Layout";
import { Price } from "../components/Price";
import { ProductImage } from "../components/ProductImage";
import { Stars } from "../components/Stars";
import { describeAttributes, formatDate } from "../format";
import { useFetch, useTitle } from "../hooks";
import type { Category, Product, RatingSummary, Review } from "../types";
import NotFound from "./NotFound";

function variantLabel(attrs: Record<string, string>, sku: string) {
  const vals = Object.values(attrs ?? {});
  return vals.length ? vals.join(" / ") : sku;
}

function BuyBox({ product, variantId }: { product: Product; variantId: number }) {
  const variant = product.variants?.find((v) => v.id === variantId);
  const { user } = useAuth();
  const { add } = useCart();
  const navigate = useNavigate();
  const [qty, setQty] = useState(1);
  const [busy, setBusy] = useState<"add" | "buy" | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [added, setAdded] = useState(false);

  useEffect(() => {
    setQty(1);
    setAdded(false);
    setError(null);
  }, [variantId]);

  if (!variant) return null;
  const available = variant.available;
  const maxQty = Math.min(available, 10);

  async function go(kind: "add" | "buy") {
    if (!user) {
      navigate(`/signin?next=${encodeURIComponent(`/p/${product.slug}`)}`);
      return;
    }
    setBusy(kind);
    setError(null);
    try {
      await add(variantId, qty);
      if (kind === "buy") navigate("/checkout");
      else setAdded(true);
    } catch (e) {
      setError(errorMessage(e));
    } finally {
      setBusy(null);
    }
  }

  return (
    <div className="panel p-4">
      <Price cents={variant.price} size="lg" />
      <p className="mt-2 text-sm text-muted">Sold and delivered by <Link to={`/s?seller=${product.seller_id}`} className="link">{product.store_name}</Link></p>
      {available > 0 ? (
        <p className={`mt-3 text-lg font-medium ${available <= 10 ? "text-bad" : "text-ok"}`}>
          {available <= 10 ? `Only ${available} licenses left` : "Available"}
        </p>
      ) : (
        <p className="mt-3 text-lg font-medium text-bad">Sold out</p>
      )}
      {available > 0 && (
        <>
          <label className="mt-3 flex items-center gap-2 text-sm">
            <span>Quantity</span>
            <select value={qty} onChange={(e) => setQty(Number(e.target.value))} className="rounded-[6px] border border-line bg-page px-2 py-1 shadow-sm">
              {Array.from({ length: maxQty }, (_, i) => i + 1).map((n) => (
                <option key={n} value={n}>{n}</option>
              ))}
            </select>
          </label>
          <div className="mt-4 flex flex-col gap-2">
            <button onClick={() => go("add")} disabled={busy !== null} className="btn-primary w-full">
              {busy === "add" ? "Adding" : "Add to cart"}
            </button>
            <button onClick={() => go("buy")} disabled={busy !== null} className="btn-buy w-full">
              {busy === "buy" ? "Adding" : "Buy now"}
            </button>
          </div>
        </>
      )}
      {added && (
        <div className="mt-3">
          <Alert tone="ok">
            Added to cart. <Link to="/cart" className="link">Go to cart</Link>
          </Alert>
        </div>
      )}
      {error && <div className="mt-3"><Alert>{error}</Alert></div>}
      <p className="mt-4 flex items-start gap-2 text-xs text-muted">
        <ShieldCheck size={16} className="shrink-0 text-ok" aria-hidden />
        After payment the seller delivers your license key or download link to your account email.
      </p>
      <p className="mt-2 text-xs text-muted">The price shown is what you pay. No fees are added at checkout.</p>
    </div>
  );
}

function ReviewForm({ productId, onDone }: { productId: number; onDone: () => void }) {
  const [rating, setRating] = useState(5);
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await api("/reviews", { method: "POST", json: { product_id: productId, rating, title, body } });
      onDone();
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={submit} className="panel mt-4 space-y-3 p-4">
      <fieldset>
        <legend className="label">Overall rating</legend>
        <div className="flex gap-1">
          {[1, 2, 3, 4, 5].map((n) => (
            <button
              type="button"
              key={n}
              onClick={() => setRating(n)}
              aria-pressed={rating === n}
              aria-label={`${n} star${n > 1 ? "s" : ""}`}
              className={`rounded-[4px] border px-2.5 py-1 text-sm tnum ${rating >= n ? "border-star bg-[#fff7e6]" : "border-line"}`}
            >
              {n}
            </button>
          ))}
        </div>
      </fieldset>
      <div>
        <label htmlFor="rv-title" className="label">Headline</label>
        <input id="rv-title" className="field" value={title} onChange={(e) => setTitle(e.target.value)} maxLength={150} placeholder="What matters most?" />
      </div>
      <div>
        <label htmlFor="rv-body" className="label">Your review</label>
        <textarea id="rv-body" className="field min-h-28" value={body} onChange={(e) => setBody(e.target.value)} maxLength={5000} />
      </div>
      {error && <Alert>{error}</Alert>}
      <button className="btn-primary" disabled={busy}>{busy ? "Submitting" : "Submit review"}</button>
    </form>
  );
}

function Reviews({ product }: { product: Product }) {
  const { user } = useAuth();
  const { data: summary, reload: reloadSummary } = useFetch<RatingSummary>(`/reviews/product/${product.id}/summary`);
  const { data: reviews, reload, setData } = useFetch<Review[]>(`/reviews/product/${product.id}`);
  const [writing, setWriting] = useState(false);
  const [voteError, setVoteError] = useState<string | null>(null);
  const mine = reviews?.some((r) => r.user_id === user?.id);

  async function helpful(id: number) {
    setVoteError(null);
    try {
      await api(`/reviews/${id}/helpful`, { method: "POST" });
      setData((rs) => rs?.map((r) => (r.id === id ? { ...r, helpful: r.helpful + 1 } : r)) ?? null);
    } catch (e) {
      setVoteError(errorMessage(e));
    }
  }

  return (
    <section id="reviews" className="mt-10 grid gap-8 border-t border-line pt-8 lg:grid-cols-[320px_1fr]">
      <div>
        <h2 className="text-[21px] font-bold">Customer reviews</h2>
        {summary && summary.count > 0 ? (
          <>
            <div className="mt-2 flex items-center gap-2">
              <Stars value={summary.average} size={18} />
              <span className="text-[17px] font-medium tnum">{summary.average.toFixed(1)} out of 5</span>
            </div>
            <p className="mt-1 text-sm text-muted">{summary.count} global ratings</p>
            <table className="mt-4 w-full text-sm">
              <tbody>
                {[5, 4, 3, 2, 1].map((s) => {
                  const n = summary.breakdown[String(s)] ?? 0;
                  const pct = summary.count ? Math.round((n / summary.count) * 100) : 0;
                  return (
                    <tr key={s}>
                      <td className="w-14 py-1 text-teal">{s} star</td>
                      <td className="py-1">
                        <div className="h-5 overflow-hidden rounded-[3px] border border-line bg-page">
                          <div className="h-full bg-star" style={{ width: `${pct}%` }} />
                        </div>
                      </td>
                      <td className="w-12 py-1 text-right tnum text-teal">{pct}%</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </>
        ) : (
          <p className="mt-2 text-sm text-muted">No reviews yet.</p>
        )}
        <div className="mt-6 border-t border-line pt-5">
          <h3 className="font-bold">Review this product</h3>
          <p className="mt-1 text-sm text-muted">Share your thoughts with other customers.</p>
          {user ? (
            !mine && (
              <button onClick={() => setWriting((w) => !w)} className="btn-quiet mt-3 w-full">
                {writing ? "Cancel" : "Write a customer review"}
              </button>
            )
          ) : (
            <Link to={`/signin?next=/p/${product.slug}`} className="btn-quiet mt-3 w-full">Sign in to write a review</Link>
          )}
          {writing && (
            <ReviewForm
              productId={product.id}
              onDone={() => {
                setWriting(false);
                reload();
                reloadSummary();
              }}
            />
          )}
        </div>
      </div>
      <div>
        {voteError && <div className="mb-3"><Alert>{voteError}</Alert></div>}
        <ul className="space-y-7">
          {(reviews ?? []).map((r) => (
            <li key={r.id}>
              <p className="text-sm font-medium">{r.author}</p>
              <div className="mt-1 flex flex-wrap items-center gap-2">
                <Stars value={r.rating} size={14} />
                {r.title && <span className="text-sm font-bold">{r.title}</span>}
              </div>
              <p className="mt-0.5 text-xs text-muted">Reviewed on {formatDate(r.created_at)}</p>
              {r.verified_purchase && (
                <p className="mt-0.5 flex items-center gap-1 text-xs font-semibold text-amber-dark"><Check size={13} aria-hidden /> Verified purchase</p>
              )}
              {r.body && <p className="mt-2 max-w-[70ch] text-[15px]">{r.body}</p>}
              <div className="mt-2 flex items-center gap-3 text-xs text-muted">
                {r.helpful > 0 && <span>{r.helpful} {r.helpful === 1 ? "person" : "people"} found this helpful</span>}
                {user && user.id !== r.user_id && (
                  <button onClick={() => helpful(r.id)} className="rounded-full border border-line px-3 py-1 text-ink hover:bg-page">Helpful</button>
                )}
              </div>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}

export default function ProductPage() {
  const { slug = "" } = useParams();
  const { data: product, loading, error } = useFetch<Product>(`/catalog/products/${encodeURIComponent(slug)}`);
  const { data: categories } = useFetch<Category[]>("/catalog/categories");
  const [variantId, setVariantId] = useState<number | null>(null);
  useTitle(product?.title ?? "");

  useEffect(() => {
    if (product?.variants?.length) {
      const firstInStock = product.variants.find((v) => v.available > 0) ?? product.variants[0];
      setVariantId(firstInStock.id);
    }
  }, [product]);

  const trail = useMemo(() => {
    const cat = categories?.find((c) => c.slug === product?.category_slug);
    const parent = cat?.parent_id ? categories?.find((c) => c.id === cat.parent_id) : undefined;
    return [parent, cat].filter(Boolean) as Category[];
  }, [categories, product]);

  if (loading) return <Page><Loading /></Page>;
  if (error && !product) return <NotFound />;
  if (!product) return null;
  const variant = product.variants?.find((v) => v.id === variantId);

  return (
    <div className="bg-white">
      <Page>
        <nav aria-label="Breadcrumb" className="mb-4 text-xs text-muted">
          <ol className="flex flex-wrap items-center gap-1">
            <li><Link to="/s" className="hover:underline">All products</Link></li>
            {trail.map((c) => (
              <li key={c.slug} className="flex items-center gap-1">
                <span aria-hidden>›</span>
                <Link to={`/s?category=${c.slug}`} className="hover:underline">{c.name}</Link>
              </li>
            ))}
          </ol>
        </nav>
        <div className="grid gap-8 lg:grid-cols-[minmax(0,5fr)_minmax(0,5fr)_minmax(260px,3fr)]">
          <div className="lg:sticky lg:top-32 lg:self-start">
            <div className="aspect-square overflow-hidden rounded-[var(--radius-panel)]">
              <ProductImage src={product.image_url} category={product.category_slug} alt={product.title} size="lg" />
            </div>
          </div>
          <div>
            <h1 className="text-[24px] font-normal leading-tight">{product.title}</h1>
            <p className="mt-1 text-sm">
              Brand: <Link to={`/s?q=${encodeURIComponent(product.brand)}`} className="link">{product.brand}</Link>
            </p>
            {product.rating_count > 0 && (
              <a href="#reviews" className="mt-1 flex items-center gap-2 text-sm">
                <span className="tnum">{product.rating_avg.toFixed(1)}</span>
                <Stars value={product.rating_avg} />
                <span className="link">{product.rating_count} ratings</span>
              </a>
            )}
            <div className="my-4 border-t border-line" />
            {variant && <Price cents={variant.price} size="lg" />}
            {product.variants && product.variants.length > 1 && (
              <div className="mt-5">
                <p className="text-sm">
                  <span className="text-muted">Option: </span>
                  <span className="font-semibold">{variant ? variantLabel(variant.attributes, variant.sku) : ""}</span>
                </p>
                <div className="mt-2 flex flex-wrap gap-2">
                  {product.variants.map((v) => (
                    <button
                      key={v.id}
                      onClick={() => setVariantId(v.id)}
                      aria-pressed={v.id === variantId}
                      className={`rounded-[6px] border px-3 py-2 text-left text-sm ${
                        v.id === variantId ? "border-teal ring-2 ring-teal/25" : "border-line hover:border-muted"
                      } ${v.available === 0 ? "text-muted line-through" : ""}`}
                    >
                      <span className="block">{variantLabel(v.attributes, v.sku)}</span>
                    </button>
                  ))}
                </div>
              </div>
            )}
            <div className="my-5 border-t border-line" />
            <h2 className="text-[17px] font-bold">About this item</h2>
            <p className="mt-2 max-w-[68ch] text-[15px] leading-relaxed">{product.description}</p>
            <table className="mt-5 text-sm">
              <tbody>
                <tr><th className="py-1 pr-6 text-left font-semibold">Brand</th><td>{product.brand}</td></tr>
                <tr><th className="py-1 pr-6 text-left font-semibold">Store</th><td>{product.store_name}</td></tr>
                {variant && <tr><th className="py-1 pr-6 text-left font-semibold">SKU</th><td className="tnum">{variant.sku}</td></tr>}
                {variant && describeAttributes(variant.attributes) && (
                  <tr><th className="py-1 pr-6 text-left font-semibold">Details</th><td>{describeAttributes(variant.attributes)}</td></tr>
                )}
              </tbody>
            </table>
          </div>
          <div className="lg:sticky lg:top-32 lg:self-start">{variantId && <BuyBox product={product} variantId={variantId} />}</div>
        </div>
        <Reviews product={product} />
      </Page>
    </div>
  );
}
