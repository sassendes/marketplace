import { useState, type FormEvent, type ReactNode } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { errorMessage } from "../api";
import { useAuth } from "../auth";
import { Alert } from "../components/Feedback";
import { useTitle } from "../hooks";

function safeNext(next: string | null) {
  return next && next.startsWith("/") && !next.startsWith("//") && !next.includes("\\") ? next : "/";
}

function Shell({ title, children, footer }: { title: string; children: ReactNode; footer: ReactNode }) {
  return (
    <div className="flex min-h-[70vh] flex-col items-center bg-white px-4 py-8">
      <Link to="/" className="mb-5 text-[26px] font-bold tracking-tight">
        sassendes<span className="text-[18px] text-amber-dark">.shop</span>
      </Link>
      <div className="w-full max-w-[350px] rounded-[8px] border border-line p-6">
        <h1 className="mb-4 text-[26px] font-normal">{title}</h1>
        {children}
      </div>
      <div className="mt-6 w-full max-w-[350px] text-center text-sm">{footer}</div>
    </div>
  );
}

export function SignIn() {
  useTitle("Sign in");
  const { login } = useAuth();
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await login(email, password);
      navigate(safeNext(params.get("next")), { replace: true });
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <Shell
      title="Sign in"
      footer={
        <>
          <p className="text-muted">New to sassendes.shop?</p>
          <Link to={`/register${params.get("next") ? `?next=${encodeURIComponent(params.get("next")!)}` : ""}`} className="btn-quiet mt-2 w-full">
            Create your account
          </Link>
        </>
      }
    >
      <form onSubmit={submit} className="space-y-4" noValidate>
        <div>
          <label htmlFor="email" className="label">Email</label>
          <input id="email" type="email" autoComplete="email" required className="field" value={email} onChange={(e) => setEmail(e.target.value)} />
        </div>
        <div>
          <label htmlFor="password" className="label">Password</label>
          <input id="password" type="password" autoComplete="current-password" required className="field" value={password} onChange={(e) => setPassword(e.target.value)} />
        </div>
        {error && <Alert>{error}</Alert>}
        <button className="btn-primary w-full rounded-[8px]" disabled={busy || !email || !password}>{busy ? "Signing in" : "Sign in"}</button>
      </form>
    </Shell>
  );
}

export function Register() {
  useTitle("Create account");
  const { register } = useAuth();
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [seller, setSeller] = useState(params.get("seller") === "1");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    if (password !== confirm) {
      setError("Passwords must match.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await register(name, email, password, seller ? "seller" : "buyer");
      navigate(seller ? "/sell" : safeNext(params.get("next")), { replace: true });
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <Shell
      title="Create account"
      footer={
        <p>
          Already have an account? <Link to="/signin" className="link">Sign in</Link>
        </p>
      }
    >
      <form onSubmit={submit} className="space-y-4" noValidate>
        <div>
          <label htmlFor="name" className="label">Your name</label>
          <input id="name" autoComplete="name" className="field" value={name} onChange={(e) => setName(e.target.value)} maxLength={50} />
        </div>
        <div>
          <label htmlFor="remail" className="label">Email</label>
          <input id="remail" type="email" autoComplete="email" className="field" value={email} onChange={(e) => setEmail(e.target.value)} />
        </div>
        <div>
          <label htmlFor="rpass" className="label">Password</label>
          <input id="rpass" type="password" autoComplete="new-password" className="field" value={password} onChange={(e) => setPassword(e.target.value)} maxLength={128} />
          <p className="mt-1 text-xs text-muted">At least 8 characters.</p>
        </div>
        <div>
          <label htmlFor="rconfirm" className="label">Re-enter password</label>
          <input id="rconfirm" type="password" autoComplete="new-password" className="field" value={confirm} onChange={(e) => setConfirm(e.target.value)} maxLength={128} />
        </div>
        <label className="flex items-start gap-2 text-sm">
          <input type="checkbox" checked={seller} onChange={(e) => setSeller(e.target.checked)} className="mt-1 accent-teal" />
          <span>I want to sell on sassendes.shop</span>
        </label>
        {error && <Alert>{error}</Alert>}
        <button className="btn-primary w-full rounded-[8px]" disabled={busy}>{busy ? "Creating your account" : "Create your account"}</button>
      </form>
    </Shell>
  );
}
