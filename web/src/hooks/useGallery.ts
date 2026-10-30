import { useEffect, useRef, useState } from "react";
import { apiURL, getJSON, type Gallery, type Item } from "../api";

export function useGallery(galleryRoute: string) {
  const [data, setData] = useState<Gallery>();
  const [items, setItems] = useState<Item[]>([]);
  const [page, setPage] = useState(1);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [retry, setRetry] = useState(0);
  const generation = useRef(0);
  const loading = useRef(false);
  const loadedPage = useRef(0);
  useEffect(() => {
    const id = ++generation.current;
    const abort = new AbortController();
    loading.current = true;
    loadedPage.current = 0;
    setBusy(true);
    setError("");
    setItems([]);
    setData(undefined);
    getJSON<Gallery>(apiURL("gallery", galleryRoute), abort.signal)
      .then((d) => {
        if (id === generation.current && !abort.signal.aborted) {
          loadedPage.current = d.page;
          setData(d);
          setItems(d.items);
          setPage(d.page);
        }
      })
      .catch((e) => {
        if (!abort.signal.aborted) setError(e.message);
      })
      .finally(() => {
        if (id === generation.current) {
          loading.current = false;
          setBusy(false);
        }
      });
    return () => abort.abort();
  }, [galleryRoute, retry]);
  async function loadMore() {
    // A queued observer callback can still capture the previous render's page
    // after a response finishes but before React commits the new page state.
    if (
      !data ||
      loading.current ||
      page !== loadedPage.current ||
      page >= data.pages
    )
      return;
    const id = generation.current;
    loading.current = true;
    setBusy(true);
    setError("");
    const u = new URL(galleryRoute, location.origin);
    u.searchParams.set("p", String(page + 1));
    try {
      const d = await getJSON<Gallery>(
        apiURL("gallery", u.href),
        new AbortController().signal,
      );
      if (id === generation.current) {
        loadedPage.current = d.page;
        setItems((old) => [...old, ...d.items]);
        setPage(d.page);
      }
    } catch (e) {
      if (id === generation.current) setError((e as Error).message);
    } finally {
      if (id === generation.current) {
        loading.current = false;
        setBusy(false);
      }
    }
  }
  function retryLoad() {
    if (data) void loadMore();
    else setRetry((n) => n + 1);
  }
  return { data, items, page, busy, error, loadMore, retryLoad };
}
