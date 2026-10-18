import { Link } from "react-router-dom";
import { Page } from "../components/Layout";
import { useTitle } from "../hooks";

export default function NotFound() {
  useTitle("Page not found");
  return (
    <Page narrow>
      <div className="panel mx-auto max-w-xl p-8">
        <h1 className="text-[22px] font-bold">We couldn't find that page</h1>
        <p className="mt-2 text-muted">The link may be broken or the item may have been removed by the seller.</p>
        <div className="mt-5 flex gap-3">
          <Link to="/" className="btn-primary">Go to the homepage</Link>
          <Link to="/s" className="btn-quiet">Browse products</Link>
        </div>
      </div>
    </Page>
  );
}
