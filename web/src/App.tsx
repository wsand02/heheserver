import React, { useEffect, useRef, useState } from "react";
import { galleryURL, plainClick } from "./api";
import { Viewer } from "./Viewer";
import { SiteHeader } from "./components/SiteHeader";
import { Breadcrumbs } from "./components/Breadcrumbs";
import { Filters } from "./components/Filters";
import { GalleryGrid } from "./components/GalleryGrid";
import { GalleryPagination } from "./components/GalleryPagination";
import { useGallery } from "./hooks/useGallery";
import { useTheme } from "./hooks/useTheme";

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
  const { data, items, page, busy, error, loadMore, retryLoad } =
    useGallery(galleryRoute);
  const background = useRef<HTMLDivElement>(null);
  const sentinel = useRef<HTMLDivElement>(null);
  const trigger = useRef<HTMLElement | null>(null);
  const [theme, setTheme] = useTheme();
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
        <SiteHeader theme={theme} setTheme={setTheme} visit={visit} />
        <Breadcrumbs parts={parts} visit={visit} />
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
            {error} <button onClick={retryLoad}>Retry</button>
          </p>
        )}
        <GalleryGrid items={items} busy={busy} visit={visit} open={open} />
        {!busy && data && !items.length && (
          <p className="empty">No matching files</p>
        )}
        {busy && <p role="status">Loading files…</p>}
        <div ref={sentinel} />
        <GalleryPagination
          data={data}
          page={page}
          busy={busy}
          params={params}
          loadMore={loadMore}
          visit={visit}
        />
        <footer className="site-footer">heheserver {data?.version}</footer>
      </div>
      {viewer && (
        <Viewer url={route} overlay={overlay} close={close} navigate={move} />
      )}
    </>
  );
}
