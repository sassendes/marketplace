import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { api } from "./api";
import { useAuth } from "./auth";
import type { Cart } from "./types";

interface CartState {
  cart: Cart | null;
  count: number;
  refresh: () => Promise<void>;
  add: (variantId: number, quantity: number) => Promise<void>;
  update: (variantId: number, quantity: number) => Promise<void>;
  remove: (variantId: number) => Promise<void>;
  reset: () => void;
}

const CartContext = createContext<CartState | null>(null);

export function CartProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth();
  const [cart, setCart] = useState<Cart | null>(null);

  const refresh = useCallback(async () => {
    if (!user) {
      setCart(null);
      return;
    }
    setCart(await api<Cart>("/cart"));
  }, [user]);

  useEffect(() => {
    refresh().catch(() => setCart(null));
  }, [refresh]);

  const add = useCallback(async (variantId: number, quantity: number) => {
    setCart(await api<Cart>("/cart/items", { method: "POST", json: { variant_id: variantId, quantity } }));
  }, []);
  const update = useCallback(async (variantId: number, quantity: number) => {
    setCart(await api<Cart>(`/cart/items/${variantId}`, { method: "PATCH", json: { quantity } }));
  }, []);
  const remove = useCallback(async (variantId: number) => {
    setCart(await api<Cart>(`/cart/items/${variantId}`, { method: "DELETE" }));
  }, []);
  const reset = useCallback(() => setCart((c) => (c ? { ...c, items: [], item_count: 0, subtotal: 0 } : c)), []);

  const value = useMemo(
    () => ({ cart, count: cart?.item_count ?? 0, refresh, add, update, remove, reset }),
    [cart, refresh, add, update, remove, reset],
  );
  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

export function useCart(): CartState {
  const ctx = useContext(CartContext);
  if (!ctx) throw new Error("useCart must be used inside CartProvider");
  return ctx;
}
