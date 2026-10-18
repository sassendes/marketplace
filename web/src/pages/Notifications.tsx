import { api } from "../api";
import { Alert, Loading } from "../components/Feedback";
import { Page } from "../components/Layout";
import { formatDate } from "../format";
import { useFetch, useTitle } from "../hooks";
import type { Notification } from "../types";

export default function Notifications() {
  useTitle("Notifications");
  const { data, loading, error, setData } = useFetch<Notification[]>("/notifications");
  const unread = data?.filter((n) => !n.read).length ?? 0;

  async function readAll() {
    await api("/notifications/read-all", { method: "POST" });
    setData((d) => d?.map((n) => ({ ...n, read: true })) ?? null);
  }

  return (
    <Page narrow>
      <div className="flex items-baseline justify-between">
        <h1 className="text-[26px] font-normal">Notifications</h1>
        {unread > 0 && <button onClick={readAll} className="link text-sm">Mark all as read</button>}
      </div>
      {loading && <Loading />}
      {error && <Alert>{error}</Alert>}
      {data && data.length === 0 && <p className="panel mt-4 p-6 text-muted">You are all caught up.</p>}
      <ul className="panel mt-4 divide-y divide-line">
        {data?.map((n) => (
          <li key={n.id} className={`px-5 py-4 ${n.read ? "" : "bg-[#f4fafb]"}`}>
            <div className="flex justify-between gap-3">
              <p className={n.read ? "" : "font-bold"}>{n.subject}</p>
              <time className="shrink-0 text-xs text-muted">{formatDate(n.created_at)}</time>
            </div>
            {n.body && <p className="mt-1 text-sm text-muted">{n.body}</p>}
          </li>
        ))}
      </ul>
    </Page>
  );
}
