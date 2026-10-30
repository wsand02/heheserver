import React, { useEffect, useRef, useState } from "react";
import {
  apiURL,
  getJSON,
  galleryURL,
  plainClick,
  type Item,
  type Gallery,
} from "./api";
import { Viewer } from "./Viewer";

function Preview({ item }: { item: Item }) {
  const [failed, setFailed] = useState(false);
  return item.thumbnail && !failed ? (
    <img
      src={item.thumbnail}
      alt=""
      loading="lazy"
      width={item.width || 640}
      height={item.height || 360}
      onError={() => setFailed(true)}
    />
  ) : (
    <span className="placeholder">
      {failed
        ? "Preview unavailable"
        : { video: "▶", audio: "♫", dir: "▰" }[item.type] || "◇"}
    </span>
  );
}
function Filters({
  url,
  apply,
  clear,
}: {
  url: string;
  apply: (form: HTMLFormElement) => void;
  clear: () => void;
}) {
  const form = useRef<HTMLFormElement>(null);
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  useEffect(() => {
    const p = new URL(url, location.origin).searchParams;
    const node = form.current!;
    for (const key of ["q", "ext"]) {
      (node.elements.namedItem(key) as HTMLInputElement).value =
        p.get(key) || "";
    }
    node
      .querySelectorAll<HTMLInputElement>('[name="type"]')
      .forEach((input) => {
        input.checked = p.getAll("type").includes(input.value);
      });
  }, [url]);
  useEffect(() => () => clearTimeout(timer.current), []);
  function submit() {
    clearTimeout(timer.current);
    if (form.current) apply(form.current);
  }
  return (
    <form
      ref={form}
      className="filters"
      onInput={() => {
        clearTimeout(timer.current);
        timer.current = setTimeout(submit, 250);
      }}
      onSubmit={(e) => {
        e.preventDefault();
        submit();
      }}
    >
      <label>
        Search
        <input name="q" placeholder="Find a filename" />
      </label>
      <details>
        <summary>Filters</summary>
        <fieldset>
          <legend>File types</legend>
          {["image", "video", "audio", "dir", "other"].map((type) => (
            <label key={type}>
              <input type="checkbox" name="type" value={type} />
              {type}
            </label>
          ))}
        </fieldset>
        <label>
          Extensions
          <input name="ext" placeholder=".png, .mp4" />
        </label>
      </details>
      <button type="submit">Apply</button>
      <button
        type="button"
        onClick={() => {
          clearTimeout(timer.current);
          clear();
        }}
      >
        Clear
      </button>
    </form>
  );
}
export function App() {
  const initial = location.pathname === "/post/";
  const [route, setRoute] = useState(location.pathname + location.search);
  const [viewer, setViewer] = useState(
    initial ? location.pathname + location.search : "",
  );
  const [overlay, setOverlay] = useState(false);
  const [galleryRoute, setGalleryRoute] = useState(
    initial ? galleryURL(location.href) : location.pathname + location.search,
  );
  const [data, setData] = useState<Gallery>();
  const [items, setItems] = useState<Item[]>([]);
  const [page, setPage] = useState(1);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [retry, setRetry] = useState(0);
  const background = useRef<HTMLDivElement>(null);
  const sentinel = useRef<HTMLDivElement>(null);
  const trigger = useRef<HTMLElement | null>(null);
  const generation = useRef(0);
  const loading = useRef(false);
  const loadedPage = useRef(0);
  const [theme, setTheme] = useState(() => {
    try {
      return localStorage.getItem("theme") || "auto";
    } catch {
      return "auto";
    }
  });
  useEffect(() => {
    const preference = matchMedia("(prefers-color-scheme: dark)");
    const update = () => {
      document.documentElement.dataset.theme =
        theme === "auto" ? (preference.matches ? "dark" : "light") : theme;
    };
    update();
    preference.addEventListener("change", update);
    try {
      localStorage.setItem("theme", theme);
    } catch {}
    return () => preference.removeEventListener("change", update);
  }, [theme]);
  useEffect(() => {
    function pop() {
      const u = location.pathname + location.search;
      setRoute(u);
      if (location.pathname === "/post/") {
        setViewer(u);
        setOverlay(Boolean(history.state?.gallery));
        if (history.state?.gallery) setGalleryRoute(history.state.gallery);
      } else {
        setViewer("");
        setOverlay(false);
        setGalleryRoute(u);
        requestAnimationFrame(() =>
          trigger.current?.focus({ preventScroll: true }),
        );
      }
    }
    window.addEventListener("popstate", pop);
    return () => window.removeEventListener("popstate", pop);
  }, []);
  useEffect(() => {
    if (background.current) background.current.inert = overlay;
  }, [overlay]);
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
  useEffect(() => {
    if (!data?.infinite || overlay || error || !sentinel.current) return;
    let active = true;
    const observer = new IntersectionObserver(
      (entries) => {
        if (active && entries.some((e) => e.isIntersecting)) void loadMore();
      },
      { rootMargin: "300px" },
    );
    observer.observe(sentinel.current);
    return () => {
      active = false;
      observer.disconnect();
    };
  }, [data, page, busy, overlay, error]);
  function visit(u: string, replace = false) {
    if (replace) history.replaceState(null, "", u);
    else history.pushState(null, "", u);
    setRoute(u);
    setGalleryRoute(u);
    setViewer("");
    setOverlay(false);
    window.scrollTo(0, 0);
  }
  function open(e: React.MouseEvent<HTMLAnchorElement>, u: string) {
    if (!plainClick(e)) return;
    e.preventDefault();
    trigger.current = e.currentTarget;
    history.pushState({ gallery: galleryRoute }, "", u);
    setRoute(u);
    setViewer(u);
    setOverlay(true);
  }
  function move(u: string) {
    // Adjacent items share one history entry: Back always returns to the grid.
    history.replaceState(overlay ? { gallery: galleryRoute } : null, "", u);
    setRoute(u);
    setViewer(u);
  }
  function close() {
    if (overlay) history.back();
    else visit(galleryURL(viewer));
  }
  function filter(form: HTMLFormElement) {
    const params = new URLSearchParams(
      new FormData(form) as unknown as URLSearchParams,
    );
    for (const [key, value] of [...params])
      if (!value.trim()) params.delete(key);
    params.set(
      "path",
      new URL(galleryRoute, location.origin).searchParams.get("path") || "/",
    );
    visit("/?" + params, true);
  }
  const params = new URL(galleryRoute, location.origin).searchParams;
  const dir = params.get("path") || "/";
  const parts = dir.split("/").filter(Boolean);
  return (
    <>
      <div
        ref={background}
        className="app"
        hidden={Boolean(viewer && !overlay)}
      >
        <header className="site-header">
          <a
            href="/"
            onClick={(e) => {
              if (plainClick(e)) {
                e.preventDefault();
                visit("/");
              }
            }}
          >
            heheserver <span>gallery</span>
          </a>
          <select
            aria-label="Theme"
            value={theme}
            onChange={(e) => setTheme(e.target.value)}
          >
            <option value="auto">System theme</option>
            <option value="light">Light</option>
            <option value="dark">Dark</option>
          </select>
        </header>
        <nav aria-label="Breadcrumb" className="breadcrumbs">
          <a
            href="/"
            onClick={(e) => {
              if (plainClick(e)) {
                e.preventDefault();
                visit("/");
              }
            }}
          >
            Home
          </a>
          {parts.map((part, i) => {
            const u =
              "/?" +
              new URLSearchParams({
                path: "/" + parts.slice(0, i + 1).join("/") + "/",
              });
            return (
              <React.Fragment key={u}>
                <span>/</span>
                <a
                  href={u}
                  onClick={(e) => {
                    if (plainClick(e)) {
                      e.preventDefault();
                      visit(u);
                    }
                  }}
                >
                  {part}
                </a>
              </React.Fragment>
            );
          })}
        </nav>
        <Filters
          url={galleryRoute}
          apply={filter}
          clear={() => visit("/?" + new URLSearchParams({ path: dir }))}
        />

        <div className="collection-heading">
          <h1>{parts.at(-1) || "Your files"}</h1>
          <span>{data ? `${data.total} items` : ""}</span>
        </div>
        {error && (
          <p role="alert">
            {error}{" "}
            <button
              onClick={() => (data ? void loadMore() : setRetry((n) => n + 1))}
            >
              Retry
            </button>
          </p>
        )}
        <main className="grid" aria-busy={busy}>
          {items.map((item) => (
            <article className="tile" key={item.path}>
              <a
                className="preview"
                aria-label={"Open " + item.name}
                href={item.type === "dir" ? item.url : item.viewerUrl}
                onClick={(e) => {
                  if (item.type === "dir") {
                    if (plainClick(e)) {
                      e.preventDefault();
                      visit(item.url);
                    }
                  } else open(e, item.viewerUrl);
                }}
              >
                <Preview item={item} />
                {item.type === "video" && <span className="play">▶</span>}
              </a>
              <div className="tile-details">
                <a
                  href={item.type === "dir" ? item.url : item.viewerUrl}
                  onClick={(e) => {
                    if (item.type === "dir") {
                      if (plainClick(e)) {
                        e.preventDefault();
                        visit(item.url);
                      }
                    } else open(e, item.viewerUrl);
                  }}
                >
                  {item.name}
                </a>
                {item.type !== "dir" && (
                  <a
                    className="download"
                    aria-label={"Download " + item.name}
                    href={item.url}
                    download
                  >
                    ↓
                  </a>
                )}
              </div>
            </article>
          ))}
        </main>
        {!busy && data && !items.length && (
          <p className="empty">No matching files</p>
        )}
        {busy && <p role="status">Loading files…</p>}
        <div ref={sentinel} />
        {data && page < data.pages && data.infinite && (
          <button disabled={busy} onClick={() => void loadMore()}>
            Load more
          </button>
        )}
        {data && !data.infinite && (
          <nav className="pagination" aria-label="Pagination">
            <button
              disabled={page <= 1}
              onClick={() => {
                const p = new URLSearchParams(params);
                p.set("p", String(page - 1));
                visit("/?" + p);
              }}
            >
              Previous
            </button>
            <span>
              {page} / {data.pages}
            </span>
            <button
              disabled={page >= data.pages}
              onClick={() => {
                const p = new URLSearchParams(params);
                p.set("p", String(page + 1));
                visit("/?" + p);
              }}
            >
              Next
            </button>
          </nav>
        )}
        <footer className="site-footer">heheserver {data?.version}</footer>
      </div>
      {viewer && (
        <Viewer url={route} overlay={overlay} close={close} navigate={move} />
      )}
    </>
  );
}
