import React, { useState } from "react";
import { plainClick, type Item } from "../api";

function DirectoryCell({ src }: { src: string }) {
  const [failed, setFailed] = useState(false);
  return (
    <span className="directory-cell">
      {failed ? (
        <span className="placeholder">▰</span>
      ) : (
        <img src={src} alt="" loading="lazy" onError={() => setFailed(true)} />
      )}
    </span>
  );
}

function Preview({ item }: { item: Item }) {
  const [failed, setFailed] = useState(false);
  if (item.type === "dir" && item.previews?.length) {
    const previews = item.previews.slice(0, 4);
    return (
      <>
        <span
          className={`directory-collage directory-collage-${previews.length}`}
        >
          {previews.map((src) => (
            <DirectoryCell key={src} src={src} />
          ))}
        </span>
        <span className="folder-badge">▰ Folder</span>
      </>
    );
  }
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
export function GalleryGrid({
  items,
  busy,
  visit,
  open,
}: {
  items: Item[];
  busy: boolean;
  visit: (url: string) => void;
  open: (event: React.MouseEvent<HTMLAnchorElement>, url: string) => void;
}) {
  return (
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
  );
}
