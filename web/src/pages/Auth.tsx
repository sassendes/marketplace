import { useState, type FormEvent, type ReactNode } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { Check, Eye, EyeOff } from "lucide-react";
import { errorMessage } from "../api";
import { useAuth } from "../auth";
import { Alert } from "../components/Feedback";
import { useTitle } from "../hooks";

function safeNext(next: string | null) {
  return next && next.startsWith("/") && !next.startsWith("//") && !next.includes("\\") ? next : "/";
}

function PasswordField({ id, label, value, onChange, autoComplete }: { id: string; label: string; value: string; onChange: (v: string) => void; autoComplete: string }) {
  const [shown, setShown] = useState(false);
  return (
    <div>
      <label htmlFor={id} className="label">{label}</label>
      <div className="relative">
        <input
          id={id}
          type={shown ? "text" : "password"}
          autoComplete={autoComplete}
          required
          className="field pr-11"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          maxLength={128}
        />
        <button
          type="button"
          onClick={() => setShown((v) => !v)}
          className="absolute inset-y-0 right-0 flex w-10 items-center justify-center text-muted hover:text-ink"
          aria-label={shown ? "Hide password" : "Show password"}
          aria-pressed={shown}
        >
          {shown ? <EyeOff size={17} aria-hidden /> : <Eye size={17} aria-hidden />}
        </button>
      </div>
    </div>
  );
}

function Shell({ title, children, footer }: { title: string; children: ReactNode; footer: ReactNode }) {
  return (
    <div className="flex min-h-[70vh] flex-col items-center bg-white px-4 py-8">
      <Link to="/" className="mb-5 text-[26px] font-bold tracking-tight">
        mercatus
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
          <p className="text-muted">New to Mercatus?</p>
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
        <PasswordField id="password" label="Password" value={password} onChange={setPassword} autoComplete="current-password" />
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
  const [seller, setSeller] = useState(params.get("seller") === "1");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    if (name.trim().length < 2) {
      setError("Enter your name, at least 2 characters, so sellers know who to deliver to.");
      return;
    }
    if ([...password].length < 8) {
      setError("Your password needs at least 8 characters.");
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
          <PasswordField id="rpass" label="Password" value={password} onChange={setPassword} autoComplete="new-password" />
          <p className={`mt-1 flex items-center gap-1 text-xs ${[...password].length >= 8 ? "text-ok" : "text-muted"}`}>
            {[...password].length >= 8 && <Check size={13} aria-hidden />}
            At least 8 characters
          </p>
        </div>
        <label className="flex items-start gap-2 text-sm">
          <input type="checkbox" checked={seller} onChange={(e) => setSeller(e.target.checked)} className="mt-1 accent-teal" />
          <span>I want to sell on Mercatus</span>
        </label>
        {error && <Alert>{error}</Alert>}
        <button className="btn-primary w-full rounded-[8px]" disabled={busy}>{busy ? "Creating your account" : "Create your account"}</button>
      </form>
    </Shell>
  );
}
