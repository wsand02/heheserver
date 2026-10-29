import React from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { App } from "./App";
import { Viewer } from "./Viewer";
import { apiURL, galleryURL, type Item } from "./api";
const item: Item = {
  name: "photo.png",
  path: "/photo.png",
  type: "image",
  url: "/fs/photo.png",
  viewerUrl: "/post/?path=%2Fphoto.png",
  size: 100,
  modified: "2026-01-01T00:00:00Z",
  width: 640,
  height: 360,
};
const collection = {
  items: [item],
  path: "/",
  page: 1,
  pages: 2,
  total: 2,
  infinite: true,
  version: "test",
};
const media = {
  item,
  position: 1,
  total: 2,
  next: "/post/?path=%2Fsecond.png",
  galleryUrl: "/?path=%2F",
};
const response = (value: unknown) =>
  Promise.resolve({ ok: true, json: () => Promise.resolve(value) } as Response);
beforeEach(() => {
  history.replaceState(null, "", "/");
});
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});
describe("routing", () => {
  it("keeps escaped paths and filters", () => {
    expect(apiURL("media", "/post/?path=%2Fc%2B%2B.png&type=image")).toBe(
      "/api/media?path=%2Fc%2B%2B.png&type=image",
    );
    expect(galleryURL("/post/?path=%2Fa%2Fb.png&q=cat&p=2")).toBe(
      "/?path=%2Fa%2F&q=cat&p=2",
    );
  });
  it("opens and closes an overlay without remounting gallery tiles", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn((url: string) =>
        response(url.startsWith("/api/gallery") ? collection : media),
      ),
    );
    render(<App />);
    const link = await screen.findByRole("link", { name: "photo.png" });
    const tile = link.closest("article");
    link.focus();
    fireEvent.click(link);
    await screen.findByRole("dialog");
    expect(document.querySelector<HTMLElement>(".app")?.inert).toBe(true);
    expect(tile?.isConnected).toBe(true);
    fireEvent.click(screen.getByRole("button", { name: "✕ Close" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
    expect(tile?.isConnected).toBe(true);
    await waitFor(() => expect(document.activeElement).toBe(link));
  });
  it("preserves modified clicks", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => response(collection)),
    );
    render(<App />);
    const link = await screen.findByRole("link", { name: "photo.png" });
    link.addEventListener("click", (e) => e.preventDefault());
    fireEvent.click(link, { ctrlKey: true });
    expect(screen.queryByRole("dialog")).toBeNull();
  });
  it("resets pagination when applying filters", async () => {
    history.replaceState(null, "", "/?path=%2F&p=2");
    const fetcher = vi.fn(() => response({ ...collection, page: 2 }));
    vi.stubGlobal("fetch", fetcher);
    render(<App />);
    await screen.findByRole("link", { name: "photo.png" });
    fireEvent.change(screen.getByPlaceholderText("Find a filename"), {
      target: { value: "cat" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Apply" }));
    await waitFor(() => expect(location.search).toContain("q=cat"));
    expect(new URLSearchParams(location.search).has("p")).toBe(false);
  });
});
describe("viewer", () => {
  it("closes only for overlay backdrop clicks", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => response(media)),
    );
    const close = vi.fn();
    const view = render(
      <Viewer url="/post/?path=x" overlay close={close} navigate={() => {}} />,
    );
    const image = await screen.findByRole("img");
    fireEvent.click(image);
    fireEvent.click(screen.getByRole("dialog"));
    expect(close).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("dialog").parentElement!);
    expect(close).toHaveBeenCalledTimes(1);
    close.mockClear();
    view.rerender(
      <Viewer
        url="/post/?path=x"
        overlay={false}
        close={close}
        navigate={() => {}}
      />,
    );
    fireEvent.click(screen.getByRole("region").parentElement!);
    expect(close).not.toHaveBeenCalled();
  });
  it("ignores stale responses after rapid navigation", async () => {
    let resolve!: (v: Response) => void;
    vi.stubGlobal(
      "fetch",
      vi.fn((url: string) =>
        url.includes("first")
          ? new Promise<Response>((r) => {
              resolve = r;
            })
          : response({ ...media, item: { ...item, name: "second.png" } }),
      ),
    );
    const props = { overlay: true, close: vi.fn(), navigate: vi.fn() };
    const view = render(<Viewer {...props} url="/post/?path=first" />);
    view.rerender(<Viewer {...props} url="/post/?path=second" />);
    await screen.findByRole("heading", { name: "second.png" });
    resolve(await response(media));
    await Promise.resolve();
    expect(screen.queryByRole("heading", { name: "photo.png" })).toBeNull();
  });
  it("supports keyboard navigation and reports failures", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => response(media)),
    );
    const navigate = vi.fn(),
      close = vi.fn();
    render(
      <Viewer url="/post/?path=x" overlay close={close} navigate={navigate} />,
    );
    await screen.findByRole("heading", { name: "photo.png" });
    fireEvent.keyDown(document, { key: "ArrowRight" });
    expect(navigate).toHaveBeenCalledWith(media.next);
    fireEvent.error(screen.getByRole("img"));
    expect(screen.getByRole("alert").textContent).toContain(
      "Preview unavailable",
    );
    fireEvent.keyDown(document, { key: "Escape" });
    expect(close).toHaveBeenCalled();
  });
  it("releases video sources on navigation and unmount", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => response({ ...media, item: { ...item, type: "video" } })),
    );
    const view = render(
      <Viewer
        url="/post/?path=x"
        overlay
        close={() => {}}
        navigate={() => {}}
      />,
    );
    await screen.findByRole("heading", { name: "photo.png" });
    const video = document.querySelector("video")!;
    expect(video.getAttribute("src")).toBe(item.url);
    view.unmount();
    expect(video.pause).toHaveBeenCalled();
    expect(video.hasAttribute("src")).toBe(false);
    expect(video.load).toHaveBeenCalled();
  });
});
