import { useEffect, useRef, useState, type FormEvent } from "react";
import { Link, NavLink, useNavigate, useSearchParams } from "react-router-dom";
import { Bell, ChevronDown, Menu, Search, ShoppingCart, X } from "lucide-react";
import { useAuth } from "../auth";
import { useCart } from "../cart";
import { useFetch } from "../hooks";
import type { Category } from "../types";

function Logo() {
  return (
    <Link to="/" className="flex shrink-0 items-baseline rounded px-1 py-1 text-white hover:ring-1 hover:ring-white/60">
      <span className="text-[22px] font-bold tracking-tight">mercatus</span>
    </Link>
  );
}

function SearchBar({ categories }: { categories: Category[] }) {
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const [q, setQ] = useState(params.get("q") ?? "");
  const [cat, setCat] = useState(params.get("category") ?? "");

  useEffect(() => {
    setQ(params.get("q") ?? "");
    setCat(params.get("category") ?? "");
  }, [params]);

  function submit(e: FormEvent) {
    e.preventDefault();
    const next = new URLSearchParams();
    if (q.trim()) next.set("q", q.trim());
    if (cat) next.set("category", cat);
    navigate(`/s?${next.toString()}`);
  }

  return (
    <form onSubmit={submit} role="search" className="flex h-10 w-full overflow-hidden rounded-[4px] bg-white focus-within:ring-[3px] focus-within:ring-saffron">
      <label htmlFor="search-category" className="sr-only">Department</label>
      <select
        id="search-category"
        value={cat}
        onChange={(e) => setCat(e.target.value)}
        className="hidden max-w-36 border-r border-line bg-page px-2 text-xs text-ink focus:outline-none sm:block"
      >
        <option value="">All</option>
        {categories.filter((c) => !c.parent_id).map((c) => (
          <option key={c.slug} value={c.slug}>{c.name}</option>
        ))}
      </select>
      <label htmlFor="search-q" className="sr-only">Search</label>
      <input
        id="search-q"
        value={q}
        onChange={(e) => setQ(e.target.value)}
        placeholder="Search software, tools and developers"
        className="min-w-0 flex-1 px-3 text-[15px] text-ink placeholder:text-muted focus:outline-none"
        autoComplete="off"
        maxLength={200}
      />
      <button type="submit" className="flex w-12 items-center justify-center bg-saffron text-ink hover:bg-saffron-dark" aria-label="Search">
        <Search size={20} strokeWidth={2.25} />
      </button>
    </form>
  );
}

function AccountMenu() {
  const { user, logout } = useAuth();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  useEffect(() => {
    function onDoc(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }
    document.addEventListener("mousedown", onDoc);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDoc);
      document.removeEventListener("keydown", onKey);
    };
  }, []);

  if (!user) {
    return (
      <Link to="/signin" className="hidden rounded px-2 py-1 text-white hover:ring-1 hover:ring-white/60 md:block">
        <span className="block text-xs text-white/80">Hello, sign in</span>
        <span className="block text-sm font-semibold">Account</span>
      </Link>
    );
  }
  return (
    <div className="relative hidden md:block" ref={ref}>
      <button
        onClick={() => setOpen((o) => !o)}
        aria-expanded={open}
        aria-haspopup="menu"
        className="rounded px-2 py-1 text-left text-white hover:ring-1 hover:ring-white/60"
      >
        <span className="block max-w-36 truncate text-xs text-white/80">Hello, {user.username.split(" ")[0]}</span>
        <span className="flex items-center gap-1 text-sm font-semibold">
          Account <ChevronDown size={14} aria-hidden />
        </span>
      </button>
      {open && (
        <div role="menu" className="panel absolute right-0 z-30 mt-1 w-56 py-2 text-sm shadow-lg">
          <MenuLink to="/orders" onPick={() => setOpen(false)}>Your orders</MenuLink>
          <MenuLink to="/notifications" onPick={() => setOpen(false)}>Notifications</MenuLink>
          {user.role !== "admin" && <MenuLink to="/sell" onPick={() => setOpen(false)}>{user.role === "seller" ? "Seller dashboard" : "Start selling"}</MenuLink>}
          {user.role === "admin" && <MenuLink to="/admin" onPick={() => setOpen(false)}>Admin console</MenuLink>}
          <div className="my-2 border-t border-line" />
          <button
            role="menuitem"
            className="block w-full px-4 py-1.5 text-left hover:bg-page"
            onClick={async () => {
              setOpen(false);
              await logout();
              navigate("/");
            }}
          >
            Sign out
          </button>
        </div>
      )}
    </div>
  );
}

function MenuLink({ to, children, onPick }: { to: string; children: string; onPick: () => void }) {
  return (
    <Link role="menuitem" to={to} onClick={onPick} className="block px-4 py-1.5 hover:bg-page">
      {children}
    </Link>
  );
}

function Unread() {
  const { user } = useAuth();
  const { data } = useFetch<{ unread: number }>(user ? "/notifications/unread-count" : null, [user?.id]);
  if (!user) return null;
  return (
    <Link to="/notifications" className="relative hidden rounded p-2 text-white hover:ring-1 hover:ring-white/60 md:block" aria-label={`Notifications, ${data?.unread ?? 0} unread`}>
      <Bell size={22} aria-hidden />
      {!!data?.unread && (
        <span className="absolute right-1 top-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-saffron px-1 text-[10px] font-bold text-ink">{data.unread}</span>
      )}
    </Link>
  );
}

export function Header() {
  const { count } = useCart();
  const { user } = useAuth();
  const { data: categories } = useFetch<Category[]>("/catalog/categories");
  const [drawer, setDrawer] = useState(false);
  const tops = (categories ?? []).filter((c) => !c.parent_id);

  return (
    <header className="sticky top-0 z-20">
      <div className="bg-ink">
        <div className="mx-auto flex max-w-[1500px] items-center gap-2 px-3 py-2 md:gap-4">
          <button className="rounded p-1 text-white md:hidden" onClick={() => setDrawer(true)} aria-label="Open menu">
            <Menu size={24} />
          </button>
          <Logo />
          <div className="hidden flex-1 md:block">
            <SearchBar categories={categories ?? []} />
          </div>
          <div className="ml-auto flex items-center gap-1 md:ml-0">
            <AccountMenu />
            {user && (
              <Link to="/orders" className="hidden rounded px-2 py-1 text-white hover:ring-1 hover:ring-white/60 lg:block">
                <span className="block text-xs text-white/80">Returns</span>
                <span className="block text-sm font-semibold">& Orders</span>
              </Link>
            )}
            <Unread />
            <Link to="/cart" className="flex items-end gap-1 rounded px-2 py-1 text-white hover:ring-1 hover:ring-white/60" aria-label={`Cart, ${count} items`}>
              <span className="relative">
                <ShoppingCart size={30} strokeWidth={1.75} aria-hidden />
                <span className="absolute left-[11px] top-[1px] w-[14px] text-center text-[13px] font-bold leading-none text-saffron tnum">{count > 99 ? "99" : count}</span>
              </span>
              <span className="hidden text-sm font-semibold sm:block">Cart</span>
            </Link>
          </div>
        </div>
        <div className="px-3 pb-2 md:hidden">
          <SearchBar categories={categories ?? []} />
        </div>
      </div>
      <nav className="bg-slate" aria-label="Departments">
        <ul className="mx-auto flex max-w-[1500px] gap-1 overflow-x-auto px-2 py-1 text-sm text-white [scrollbar-width:none]">
          <li>
            <NavLink to="/s" className="block whitespace-nowrap rounded px-2 py-1 hover:ring-1 hover:ring-white/60">All products</NavLink>
          </li>
          {tops.map((c) => (
            <li key={c.slug}>
              <Link to={`/s?category=${c.slug}`} className="block whitespace-nowrap rounded px-2 py-1 hover:ring-1 hover:ring-white/60">
                {c.name}
              </Link>
            </li>
          ))}
          <li>
            <Link to="/s?sort=price_asc" className="block whitespace-nowrap rounded px-2 py-1 hover:ring-1 hover:ring-white/60">Lowest prices</Link>
          </li>
          <li>
            <Link to={user ? "/sell" : "/register?seller=1"} className="block whitespace-nowrap rounded px-2 py-1 hover:ring-1 hover:ring-white/60">Sell on Mercatus</Link>
          </li>
        </ul>
      </nav>
      {drawer && <Drawer categories={categories ?? []} onClose={() => setDrawer(false)} />}
    </header>
  );
}

function Drawer({ categories, onClose }: { categories: Category[]; onClose: () => void }) {
  const { user, logout } = useAuth();
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && onClose();
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [onClose]);
  return (
    <div className="fixed inset-0 z-40 md:hidden" role="dialog" aria-modal="true" aria-label="Menu">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="absolute inset-y-0 left-0 w-[82%] max-w-80 overflow-y-auto bg-white">
        <div className="flex items-center justify-between bg-slate px-4 py-3 text-white">
          <span className="font-semibold">{user ? `Hello, ${user.username.split(" ")[0]}` : "Hello, sign in"}</span>
          <button onClick={onClose} aria-label="Close menu"><X size={22} /></button>
        </div>
        <div className="p-2 text-[15px]" onClick={onClose}>
          {!user && <Link to="/signin" className="block rounded px-3 py-2 hover:bg-page">Sign in</Link>}
          {user && <Link to="/orders" className="block rounded px-3 py-2 hover:bg-page">Your orders</Link>}
          {user && <Link to="/notifications" className="block rounded px-3 py-2 hover:bg-page">Notifications</Link>}
          {user?.role === "admin" && <Link to="/admin" className="block rounded px-3 py-2 hover:bg-page">Admin console</Link>}
          {user && user.role !== "admin" && <Link to="/sell" className="block rounded px-3 py-2 hover:bg-page">Seller dashboard</Link>}
          <div className="mx-3 my-2 border-t border-line" />
          <p className="px-3 pb-1 text-sm font-semibold">Shop by department</p>
          {categories.filter((c) => !c.parent_id).map((c) => (
            <Link key={c.slug} to={`/s?category=${c.slug}`} className="block rounded px-3 py-2 hover:bg-page">{c.name}</Link>
          ))}
          {user && (
            <>
              <div className="mx-3 my-2 border-t border-line" />
              <button onClick={() => logout()} className="block w-full rounded px-3 py-2 text-left hover:bg-page">Sign out</button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
