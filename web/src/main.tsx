import { StrictMode, Suspense, lazy } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import "@fontsource-variable/public-sans";
import "./index.css";
import { AuthProvider } from "./auth";
import { CartProvider } from "./cart";
import { Layout, Page, RequireAuth } from "./components/Layout";
import { Loading } from "./components/Feedback";
import Home from "./pages/Home";
import Search from "./pages/Search";
import ProductPage from "./pages/Product";
import CartPage from "./pages/Cart";
import NotFound from "./pages/NotFound";
import { Register, SignIn } from "./pages/Auth";

const Checkout = lazy(() => import("./pages/Checkout"));
const Orders = lazy(() => import("./pages/Orders").then((m) => ({ default: m.OrdersPage })));
const OrderDetail = lazy(() => import("./pages/Orders").then((m) => ({ default: m.OrderDetailPage })));
const Notifications = lazy(() => import("./pages/Notifications"));
const Seller = lazy(() => import("./pages/Seller"));
const Admin = lazy(() => import("./pages/Admin"));

const fallback = <Page><Loading /></Page>;

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <AuthProvider>
        <CartProvider>
          <Suspense fallback={fallback}>
            <Routes>
              <Route path="/signin" element={<SignIn />} />
              <Route path="/register" element={<Register />} />
              <Route element={<Layout />}>
                <Route index element={<Home />} />
                <Route path="s" element={<Search />} />
                <Route path="p/:slug" element={<ProductPage />} />
                <Route path="cart" element={<CartPage />} />
                <Route path="checkout" element={<RequireAuth><Checkout /></RequireAuth>} />
                <Route path="orders" element={<RequireAuth><Orders /></RequireAuth>} />
                <Route path="orders/:id" element={<RequireAuth><OrderDetail /></RequireAuth>} />
                <Route path="notifications" element={<RequireAuth><Notifications /></RequireAuth>} />
                <Route path="sell" element={<RequireAuth role={["buyer", "seller"]}><Seller /></RequireAuth>} />
                <Route path="admin" element={<RequireAuth role="admin"><Admin /></RequireAuth>} />
                <Route path="*" element={<NotFound />} />
              </Route>
            </Routes>
          </Suspense>
        </CartProvider>
      </AuthProvider>
    </BrowserRouter>
  </StrictMode>,
);
