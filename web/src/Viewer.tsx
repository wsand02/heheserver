import React, { useEffect, useLayoutEffect, useRef, useState } from "react";
import { apiURL, getJSON, type Media } from "./api";
import { Player } from "./components/Player";

export function Viewer({
  url,
  overlay,
  close,
  navigate,
}: {
  url: string;
  overlay: boolean;
  close: () => void;
  navigate: (u: string) => void;
}) {
  const [data, setData] = useState<Media>();
  const [error, setError] = useState("");
  const [retry, setRetry] = useState(0);
  const box = useRef<HTMLDivElement>(null);
  const swipe = useRef<{
    x: number;
    y: number;
    id: number;
    time: number;
  } | null>(null);
  useEffect(() => {
    swipe.current = null;
    const abort = new AbortController();
    setData(undefined);
    setError("");
    getJSON<Media>(apiURL("media", url), abort.signal)
      .then((d) => {
        if (!abort.signal.aborted) setData(d);
      })
      .catch((e) => {
        if (!abort.signal.aborted) setError(e.message);
      });
    return () => abort.abort();
  }, [url, retry]);
  useLayoutEffect(() => {
    const previous = document.activeElement as HTMLElement;
    box.current?.focus();
    const old = document.body.style.overflow;
    if (overlay) document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = old;
      if (previous?.isConnected) previous.focus({ preventScroll: true });
    };
  }, [overlay]);
  useEffect(() => {
    function key(e: KeyboardEvent) {
      if (e.key === "Escape") {
        e.preventDefault();
        close();
        return;
      }
      if (overlay && e.key === "Tab") {
        const nodes = Array.from(
          box.current?.querySelectorAll<HTMLElement>(
            "a[href],button:not(:disabled),input:not(:disabled),video,audio",
          ) || [],
        );
        const first = nodes[0],
          last = nodes[nodes.length - 1];
        if (
          e.shiftKey &&
          (document.activeElement === first ||
            document.activeElement === box.current)
        ) {
          e.preventDefault();
          last?.focus();
        } else if (
          !e.shiftKey &&
          (document.activeElement === last ||
            document.activeElement === box.current)
        ) {
          e.preventDefault();
          first?.focus();
        }
      }
      if (
        e.target instanceof Element &&
        e.target.closest("input,textarea,select,video,audio,[contenteditable]")
      )
        return;
      if (e.altKey || e.ctrlKey || e.metaKey || e.shiftKey) return;
      if (e.key === "ArrowLeft" && data?.previous) {
        e.preventDefault();
        navigate(data.previous);
      }
      if (e.key === "ArrowRight" && data?.next) {
        e.preventDefault();
        navigate(data.next);
      }
    }
    document.addEventListener("keydown", key);
    return () => document.removeEventListener("keydown", key);
  }, [data, close, navigate, overlay]);
  const nav = (target?: string) => {
    if (target) navigate(target);
  };
  return (
    <div
      className={overlay ? "viewer-backdrop" : "standalone"}
      onClick={(event) => {
        if (overlay && event.target === event.currentTarget) close();
      }}
    >
      <div
        className="viewer"
        role={overlay ? "dialog" : "region"}
        aria-modal={overlay || undefined}
        aria-label="Media viewer"
        tabIndex={-1}
        ref={box}
      >
        <header className="viewer-toolbar">
          <button onClick={close}>{overlay ? "✕ Close" : "← Gallery"}</button>
          <span aria-live="polite">
            {data?.position ? `${data.position} / ${data.total}` : ""}
          </span>
          <div>
            <button
              aria-label="Previous item"
              disabled={!data?.previous}
              onClick={() => nav(data?.previous)}
            >
              ←
            </button>
            <button
              aria-label="Next item"
              disabled={!data?.next}
              onClick={() => nav(data?.next)}
            >
              →
            </button>
          </div>
        </header>
        <div
          className={`stage${data?.item.type === "image" ? " stage-image" : ""}`}
          onTouchStart={(event) => {
            swipe.current = null;
            if (
              !data ||
              event.touches.length !== 1 ||
              (window.visualViewport?.scale ?? 1) > 1 ||
              (event.target as Element).closest(
                'audio,a,button,input,[data-zoomed="true"]',
              )
            )
              return;
            const touch = event.touches[0];
            const video = (event.target as Element).closest("video");
            // Native controls live in browser shadow DOM, so reserve their
            // bottom strip rather than treating the entire video as a control.
            if (
              video &&
              (video.matches(":fullscreen") ||
                touch.clientY >= video.getBoundingClientRect().bottom - 80)
            )
              return;
            swipe.current = {
              x: touch.clientX,
              y: touch.clientY,
              id: touch.identifier,
              time: Date.now(),
            };
          }}
          onTouchMove={(event) => {
            const start = swipe.current;
            if (!start) return;
            const touch = event.touches[0];
            if (
              event.touches.length !== 1 ||
              touch.identifier !== start.id ||
              Math.abs(touch.clientY - start.y) > 30
            )
              swipe.current = null;
          }}
          onTouchCancel={() => {
            swipe.current = null;
          }}
          onTouchEnd={(event) => {
            const start = swipe.current;
            swipe.current = null;
            if (
              !start ||
              event.touches.length ||
              Date.now() - start.time > 700 ||
              (window.visualViewport?.scale ?? 1) > 1
            )
              return;
            const touch = Array.from(event.changedTouches).find(
              (t) => t.identifier === start.id,
            );
            if (!touch) return;
            const dx = touch.clientX - start.x;
            const dy = touch.clientY - start.y;
            if (
              Math.abs(dx) < 60 ||
              Math.abs(dy) > 30 ||
              Math.abs(dx) < Math.abs(dy) * 2
            )
              return;
            nav(dx < 0 ? data?.next : data?.previous);
          }}
        >
          {error ? (
            <p role="alert">
              {error}{" "}
              <button onClick={() => setRetry((n) => n + 1)}>Retry</button>
            </p>
          ) : data ? (
            <Player key={data.item.path} item={data.item} />
          ) : (
            <p role="status">Loading media…</p>
          )}
        </div>
        {data && (data.previous || data.next) && (
          <p className="viewer-hint">
            Swipe left or right, or use the arrow keys to browse
          </p>
        )}
        {data && (
          <footer className="media-details">
            <div>
              <h2>{data.item.name}</h2>
              <small>
                {(data.item.size / 1e6).toFixed(2)} MB ·{" "}
                {new Date(data.item.modified).toLocaleString()}
              </small>
            </div>
            <a href={data.item.url} download>
              Download
            </a>
            <a href={data.item.url} target="_blank" rel="noreferrer">
              Open original
            </a>
          </footer>
        )}
      </div>
    </div>
  );
}
