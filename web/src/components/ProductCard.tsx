import { Link } from "react-router-dom";
import type { Product } from "../types";
import { Price } from "./Price";
import { ProductImage } from "./ProductImage";
import { Stars } from "./Stars";

export function ProductCard({ product }: { product: Product }) {
  const low = product.available > 0 && product.available <= 10;
  return (
    <article className="group flex h-full flex-col bg-white p-3">
      <Link to={`/p/${product.slug}`} className="block aspect-square overflow-hidden rounded-[3px]" tabIndex={-1} aria-hidden>
        <ProductImage src={product.image_url} category={product.category_slug} alt={product.title} />
      </Link>
      <div className="mt-3 flex flex-1 flex-col gap-1">
        <h3 className="line-clamp-2 text-[15px] leading-snug">
          <Link to={`/p/${product.slug}`} className="hover:text-amber-dark">
            {product.title}
          </Link>
        </h3>
        {product.rating_count > 0 ? (
          <div className="flex items-center gap-1 text-sm">
            <Stars value={product.rating_avg} size={14} />
            <span className="text-teal">{product.rating_count}</span>
          </div>
        ) : (
          <div className="text-xs text-muted">No reviews yet</div>
        )}
        <div className="mt-1">
          <Price cents={product.price_from} />
        </div>
        {product.available === 0 ? (
          <div className="text-sm text-bad">Sold out</div>
        ) : low ? (
          <div className="text-sm text-bad">Only {product.available} licenses left</div>
        ) : (
          <div className="text-xs text-muted">Sold by {product.store_name}</div>
        )}
      </div>
    </article>
  );
}

export function ProductGrid({ products }: { products: Product[] }) {
  return (
    <ul className="grid grid-cols-2 border-l border-t border-line sm:grid-cols-3 lg:grid-cols-4">
      {products.map((p) => (
        <li key={p.id} className="border-b border-r border-line">
          <ProductCard product={p} />
        </li>
      ))}
    </ul>
  );
}
