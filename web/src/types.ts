export type Role = "buyer" | "seller" | "admin";

export interface User {
  id: number;
  username: string;
  email: string;
  role: Role;
  email_verified: boolean;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

export interface Variant {
  id: number;
  product_id: number;
  sku: string;
  price: number;
  currency: string;
  available: number;
  attributes: Record<string, string>;
}

export interface Product {
  id: number;
  seller_id: number;
  store_name: string;
  title: string;
  slug: string;
  description: string;
  category_id: number | null;
  category_slug: string;
  brand: string;
  image_url: string;
  status: string;
  price_from: number;
  currency: string;
  rating_avg: number;
  rating_count: number;
  available: number;
  created_at: string;
  variants?: Variant[];
}

export interface ProductList {
  products: Product[];
  total: number;
  page: number;
  page_size: number;
}

export interface Category {
  id: number;
  parent_id: number | null;
  name: string;
  slug: string;
  product_count: number;
}

export interface CartItem {
  id: number;
  variant_id: number;
  product_id: number;
  seller_id: number;
  store_name: string;
  title: string;
  slug: string;
  image_url: string;
  category_slug: string;
  sku: string;
  attributes: Record<string, string> | null;
  quantity: number;
  unit_price: number;
  currency: string;
  available: number;
  purchasable: boolean;
}

export interface Cart {
  id: number;
  items: CartItem[];
  item_count: number;
  subtotal: number;
  currency: string;
}

export interface Review {
  id: number;
  product_id: number;
  user_id: number;
  author: string;
  rating: number;
  title: string;
  body: string;
  verified_purchase: boolean;
  helpful: number;
  created_at: string;
}

export interface RatingSummary {
  average: number;
  count: number;
  breakdown: Record<string, number>;
}

export interface OrderItem {
  id: number;
  variant_id: number;
  product_id: number;
  title: string;
  slug: string;
  image_url: string;
  category_slug: string;
  quantity: number;
  unit_price_cents: number;
}

export interface SubOrder {
  id: number;
  order_id: number;
  seller_id: number;
  store_name: string;
  status: string;
  subtotal_cents: number;
  commission_cents: number;
  created_at?: string;
  buyer_email?: string;
  summary?: string;
  items?: OrderItem[];
}

export interface Order {
  id: number;
  status: string;
  total_cents: number;
  currency: string;
  item_count: number;
  created_at: string;
  preview?: { title: string; image_url: string; slug: string; category_slug: string }[];
  sub_orders?: SubOrder[];
}

export interface CheckoutResult {
  saga_id: number;
  order_id: number;
  status: string;
  error?: string;
}

export interface Seller {
  id: number;
  store_name: string;
  slug: string;
  description: string;
  status: string;
  commission_rate: number;
  payout_enabled: boolean;
}

export interface Notification {
  id: number;
  subject: string;
  body: string;
  read: boolean;
  created_at: string;
}
