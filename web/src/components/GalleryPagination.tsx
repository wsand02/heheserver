import React from "react";
import type { Gallery } from "../api";

export function GalleryPagination({
  data,
  page,
  busy,
  params,
  loadMore,
  visit,
}: {
  data: Gallery | undefined;
  page: number;
  busy: boolean;
  params: URLSearchParams;
  loadMore: () => Promise<void>;
  visit: (url: string) => void;
}) {
  return (
    <>
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
    </>
  );
}
