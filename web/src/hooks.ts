import { useEffect, useState } from "react";
import { api, errorMessage } from "./api";

export function useFetch<T>(path: string | null, deps: unknown[] = []) {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(path !== null);
  const [tick, setTick] = useState(0);

  useEffect(() => {
    if (path === null) {
      setLoading(false);
      return;
    }
    let live = true;
    setLoading(true);
    setError(null);
    api<T>(path)
      .then((d) => live && setData(d))
      .catch((e) => live && setError(errorMessage(e)))
      .finally(() => live && setLoading(false));
    return () => {
      live = false;
    };
  }, [path, tick, ...deps]);

  return { data, error, loading, reload: () => setTick((t) => t + 1), setData };
}

export function useTitle(title: string) {
  useEffect(() => {
    document.title = title ? `${title} | sassendes.shop` : "sassendes.shop";
  }, [title]);
}
