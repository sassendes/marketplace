import { Link } from "react-router-dom";
import { ProductImage } from "../components/ProductImage";
import { ProductCard } from "../components/ProductCard";
import { SkeletonGrid } from "../components/Feedback";
import { Page } from "../components/Layout";
import { useAuth } from "../auth";
import { useFetch, useTitle } from "../hooks";
import type { Category, Product, ProductList } from "../types";
import { formatMoney } from "../format";

function DepartmentCard({ category }: { category: Category }) {
  const { data } = useFetch<ProductList>(`/catalog/products?category=${category.slug}&page_size=4&sort=rating`);
  const items = data?.products ?? [];
  return (
    <section className="flex flex-col bg-white p-5">
      <h2 className="text-[19px] font-bold">{category.name}</h2>
      <div className="mt-3 grid flex-1 grid-cols-2 gap-x-3 gap-y-2">
        {items.map((p) => (
          <Link key={p.id} to={`/p/${p.slug}`} className="group block">
            <div className="aspect-square overflow-hidden rounded-[3px]">
              <ProductImage src={p.image_url} category={p.category_slug} alt={p.title} size="sm" />
            </div>
            <p className="mt-1 line-clamp-1 text-xs group-hover:text-amber-dark">{p.title}</p>
            <p className="tnum text-xs text-muted">from {formatMoney(p.price_from)}</p>
          </Link>
        ))}
        {items.length === 0 && Array.from({ length: 4 }).map((_, i) => <div key={i} className="aspect-square animate-pulse rounded bg-page" />)}
      </div>
      <Link to={`/s?category=${category.slug}`} className="link mt-4 text-sm">
        See all {category.product_count} in {category.name}
      </Link>
    </section>
  );
}

function Shelf({ title, query, href }: { title: string; query: string; href: string }) {
  const { data, loading } = useFetch<ProductList>(`/catalog/products?${query}`);
  return (
    <section className="mt-5 bg-white p-5">
      <div className="mb-3 flex items-baseline gap-4">
        <h2 className="text-[19px] font-bold">{title}</h2>
        <Link to={href} className="link text-sm">See more</Link>
      </div>
      {loading ? (
        <SkeletonGrid count={4} />
      ) : (
        <ul className="-mx-1 flex snap-x gap-0 overflow-x-auto pb-2">
          {(data?.products ?? []).map((p: Product) => (
            <li key={p.id} className="w-[46%] shrink-0 snap-start sm:w-[30%] md:w-[22%] lg:w-[16.6%]">
              <ProductCard product={p} />
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

function Welcome() {
  const { user } = useAuth();
  return (
    <div className="bg-slate text-white">
      <div className="mx-auto flex max-w-[1500px] flex-col gap-4 px-5 py-8 md:flex-row md:items-end md:justify-between">
        <div className="max-w-xl">
          <h1 className="text-[26px] font-bold leading-tight md:text-[30px]">
            {user ? `Welcome back, ${user.username.split(" ")[0]}` : "Software from independent developers"}
          </h1>
          <p className="mt-2 text-white/80">
            Apps, tools, games and license keys sold directly by the people who make them. Pay once at checkout, each seller delivers your software.
          </p>
        </div>
        {!user && (
          <div className="flex gap-3">
            <Link to="/signin" className="btn-primary px-6">Sign in</Link>
            <Link to="/register" className="btn border border-white/40 px-6 text-white hover:bg-white/10">Create account</Link>
          </div>
        )}
      </div>
    </div>
  );
}

export default function Home() {
  useTitle("");
  const { data: categories } = useFetch<Category[]>("/catalog/categories");
  const tops = (categories ?? []).filter((c) => !c.parent_id && c.product_count > 0).slice(0, 4);
  return (
    <>
      <Welcome />
      <Page>
        <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
          {tops.map((c) => (
            <DepartmentCard key={c.slug} category={c} />
          ))}
        </div>

        <Shelf title="New arrivals" query="sort=newest&page_size=12" href="/s?sort=newest" />

      </Page>
    </>
  );
}
