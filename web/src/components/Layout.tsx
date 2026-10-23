import { Link, Navigate, Outlet, useLocation } from "react-router-dom";
import { useEffect, type ReactNode } from "react";
import { Header } from "./Header";
import { useAuth } from "../auth";
import { Loading } from "./Feedback";
import type { Role } from "../types";

function Footer() {
  return (
    <footer className="mt-12 bg-slate text-sm text-white/85">
      <button
        onClick={() => window.scrollTo({ top: 0 })}
        className="block w-full bg-[#37475a] py-3 text-center text-[13px] text-white hover:bg-[#435469]"
      >
        Back to top
      </button>
      <div className="mx-auto grid max-w-[1100px] grid-cols-2 gap-8 px-6 py-10 md:grid-cols-4">
        <div>
          <p className="mb-2 font-semibold text-white">Shop</p>
          <ul className="space-y-1.5">
            <li><Link className="hover:underline" to="/s">All products</Link></li>
            <li><Link className="hover:underline" to="/s?sort=rating">Top rated</Link></li>
            <li><Link className="hover:underline" to="/s?sort=newest">New arrivals</Link></li>
          </ul>
        </div>
        <div>
          <p className="mb-2 font-semibold text-white">Sell</p>
          <ul className="space-y-1.5">
            <li><Link className="hover:underline" to="/sell">Open a store</Link></li>
            <li><Link className="hover:underline" to="/sell">Seller dashboard</Link></li>
          </ul>
        </div>
        <div>
          <p className="mb-2 font-semibold text-white">Your account</p>
          <ul className="space-y-1.5">
            <li><Link className="hover:underline" to="/orders">Your orders</Link></li>
            <li><Link className="hover:underline" to="/cart">Cart</Link></li>
          </ul>
        </div>
        <div>
          <p className="mb-2 font-semibold text-white">About Mercatus</p>
          <p className="text-white/70">A software marketplace where independent sellers list apps, tools and license keys and deliver them straight to buyers.</p>
          <ul className="mt-3 space-y-1.5">
            <li><a className="hover:underline" href="https://github.com/sassendes/mercatus" target="_blank" rel="noreferrer">Source code</a></li>
            <li><Link className="hover:underline" to="/p/seller-guide">Seller guide</Link></li>
          </ul>
        </div>
      </div>
      <div className="border-t border-white/10 bg-ink py-5 text-center text-xs text-white/60">
        <p>© {new Date().getFullYear()} Mercatus. Prices shown are final, with no fees added at checkout.</p>
        <p className="mt-1">Payments run in test mode. No real money moves.</p>
      </div>
    </footer>
  );
}

export function Layout() {
  const { pathname } = useLocation();
  useEffect(() => {
    window.scrollTo(0, 0);
  }, [pathname]);
  return (
    <div className="flex min-h-screen flex-col">
      <a href="#main" className="sr-only focus:not-sr-only focus:absolute focus:left-2 focus:top-2 focus:z-50 focus:rounded focus:bg-white focus:px-3 focus:py-2">
        Skip to content
      </a>
      <Header />
      <main id="main" className="flex-1">
        <Outlet />
      </main>
      <Footer />
    </div>
  );
}

export function RequireAuth({ children, role }: { children: ReactNode; role?: Role | Role[] }) {
  const { user, loading } = useAuth();
  const location = useLocation();
  if (loading) return <div className="mx-auto max-w-[1500px] px-4"><Loading /></div>;
  if (!user) return <Navigate to={`/signin?next=${encodeURIComponent(location.pathname + location.search)}`} replace />;
  const roles = role ? (Array.isArray(role) ? role : [role]) : null;
  if (roles && !roles.includes(user.role)) return <Navigate to="/" replace />;
  return <>{children}</>;
}

export function Page({ children, narrow }: { children: ReactNode; narrow?: boolean }) {
  return <div className={`mx-auto px-3 py-5 md:px-5 ${narrow ? "max-w-[1100px]" : "max-w-[1500px]"}`}>{children}</div>;
}
