import { Link, useSearchParams } from "react-router-dom";
import { ProductGrid } from "../components/ProductCard";
import { Alert, SkeletonGrid, RetryButton } from "../components/Feedback";
import { Page } from "../components/Layout";
import { useFetch, useTitle } from "../hooks";
import type { Category, ProductList } from "../types";

const priceRanges = [
  { label: "Under $25", min: "", max: "2500" },
  { label: "$25 to $50", min: "2500", max: "5000" },
  { label: "$50 to $100", min: "5000", max: "10000" },
  { label: "$100 to $200", min: "10000", max: "20000" },
  { label: "$200 and above", min: "20000", max: "" },
];

const sorts = [
  { value: "", label: "Featured" },
  { value: "price_asc", label: "Price: low to high" },
  { value: "price_desc", label: "Price: high to low" },
  { value: "rating", label: "Customer rating" },
  { value: "newest", label: "Newest arrivals" },
];

export default function Search() {
  const [params, setParams] = useSearchParams();
  const q = params.get("q") ?? "";
  const category = params.get("category") ?? "";
  const page = Math.max(1, Number(params.get("page") ?? 1));
  const { data: categories } = useFetch<Category[]>("/catalog/categories");

  const query = new URLSearchParams();
  if (q) query.set("search", q);
  if (category) query.set("category", category);
  if (params.get("seller")) query.set("seller_id", params.get("seller")!);
  for (const k of ["min_price", "max_price", "sort"]) {
    const v = params.get(k);
    if (v) query.set(k, v);
  }
  query.set("page", String(page));
  query.set("page_size", "24");
  const { data, loading, error, reload } = useFetch<ProductList>(`/catalog/products?${query.toString()}`);

  const current = categories?.find((c) => c.slug === category);
  useTitle(q ? `Results for ${q}` : current?.name ?? "All products");

  function setParam(key: string, value: string) {
    const next = new URLSearchParams(params);
    if (value) next.set(key, value);
    else next.delete(key);
    next.delete("page");
    setParams(next);
  }

  function setPrice(min: string, max: string) {
    const next = new URLSearchParams(params);
    min ? next.set("min_price", min) : next.delete("min_price");
    max ? next.set("max_price", max) : next.delete("max_price");
    next.delete("page");
    setParams(next);
  }

  const parent = current?.parent_id ? categories?.find((c) => c.id === current.parent_id) : current;
  const children = parent ? (categories ?? []).filter((c) => c.parent_id === parent.id) : [];
  const tops = (categories ?? []).filter((c) => !c.parent_id);
  const totalPages = data ? Math.max(1, Math.ceil(data.total / data.page_size)) : 1;
  const activeMin = params.get("min_price") ?? "";
  const activeMax = params.get("max_price") ?? "";

  return (
    <>
      <div className="border-b border-line bg-white">
        <div className="mx-auto flex max-w-[1500px] flex-wrap items-center justify-between gap-3 px-3 py-2 md:px-5">
          <p className="text-sm">
            {loading ? "Searching" : `${data?.total ?? 0} ${data?.total === 1 ? "result" : "results"}`}
            {q && (
              <>
                {" "}for <span className="font-semibold text-amber-dark">"{q}"</span>
              </>
            )}
            {current && <> in <span className="font-semibold">{current.name}</span></>}
          </p>
          <label className="flex items-center gap-2 text-sm">
            <span className="text-muted">Sort by</span>
            <select
              value={params.get("sort") ?? ""}
              onChange={(e) => setParam("sort", e.target.value)}
              className="rounded-[6px] border border-line bg-page px-2 py-1 text-sm shadow-sm"
            >
              {sorts.map((s) => (
                <option key={s.value} value={s.value}>{s.label}</option>
              ))}
            </select>
          </label>
        </div>
      </div>
      <Page>
        <div className="flex flex-col gap-6 md:flex-row">
          <aside className="w-full shrink-0 text-sm md:w-56" aria-label="Filters">
            <h2 className="font-bold">Department</h2>
            <ul className="mt-2 space-y-1">
              <li>
                <button onClick={() => setParam("category", "")} className={!category ? "font-semibold" : "hover:text-amber-dark"}>
                  Any department
                </button>
              </li>
              {(parent ? [parent] : tops).map((c) => (
                <li key={c.slug}>
                  <button onClick={() => setParam("category", c.slug)} className={c.slug === category ? "font-semibold" : "hover:text-amber-dark"}>
                    {c.name}
                  </button>
                  {parent && c.id === parent.id && children.length > 0 && (
                    <ul className="mt-1 space-y-1 pl-3">
                      {children.map((ch) => (
                        <li key={ch.slug}>
                          <button onClick={() => setParam("category", ch.slug)} className={ch.slug === category ? "font-semibold" : "hover:text-amber-dark"}>
                            {ch.name} <span className="text-muted">({ch.product_count})</span>
                          </button>
                        </li>
                      ))}
                    </ul>
                  )}
                </li>
              ))}
            </ul>
            <h2 className="mt-6 font-bold">Price</h2>
            <ul className="mt-2 space-y-1">
              {priceRanges.map((r) => {
                const active = r.min === activeMin && r.max === activeMax;
                return (
                  <li key={r.label}>
                    <button onClick={() => (active ? setPrice("", "") : setPrice(r.min, r.max))} className={active ? "font-semibold" : "hover:text-amber-dark"}>
                      {r.label}
                    </button>
                  </li>
                );
              })}
            </ul>
            {(activeMin || activeMax || category || q) && (
              <Link to="/s" className="link mt-6 inline-block">Clear all filters</Link>
            )}
          </aside>
          <section className="min-w-0 flex-1" aria-live="polite">
            {error && <Alert action={<RetryButton onClick={reload} />}>Products could not be loaded. {error}</Alert>}
            {loading ? (
              <SkeletonGrid />
            ) : data && data.products.length > 0 ? (
              <>
                <ProductGrid products={data.products} />
                {totalPages > 1 && (
                  <nav className="mt-6 flex items-center justify-center gap-1" aria-label="Pagination">
                    <button
                      disabled={page <= 1}
                      onClick={() => setParams((p) => { const n = new URLSearchParams(p); n.set("page", String(page - 1)); return n; })}
                      className="btn-quiet rounded-[6px]"
                    >
                      Previous
                    </button>
                    <span className="px-3 text-sm tnum">Page {page} of {totalPages}</span>
                    <button
                      disabled={page >= totalPages}
                      onClick={() => setParams((p) => { const n = new URLSearchParams(p); n.set("page", String(page + 1)); return n; })}
                      className="btn-quiet rounded-[6px]"
                    >
                      Next
                    </button>
                  </nav>
                )}
              </>
            ) : (
              !error && (
                <div className="panel p-8">
                  <h2 className="text-lg font-bold">{q ? `No results for "${q}"` : "No products match these filters"}</h2>
                  <p className="mt-1 text-sm text-muted">Check the spelling, try a more general term, or remove a filter.</p>
                  <div className="mt-4 flex flex-wrap gap-2">
                    {q && (activeMin || activeMax || category) && (
                      <Link to={`/s?q=${encodeURIComponent(q)}`} className="btn-quiet">Keep the search, clear filters</Link>
                    )}
                    <Link to="/s" className="btn-primary">Browse all products</Link>
                  </div>
                </div>
              )
            )}
          </section>
        </div>
      </Page>
    </>
  );
}
