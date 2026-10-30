import React from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  cleanup,
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { App } from "./App";
import { Viewer } from "./Viewer";
import { Player } from "./components/Player";
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
it("zooms video toward the cursor, pans, and keeps playback controls outside the viewport", () => {
  render(<Player item={{ ...item, type: "video" }} />);
  const video = document.querySelector("video")!;
  const viewport = video.parentElement!;
  Object.defineProperties(video, {
    videoWidth: { value: 640 },
    videoHeight: { value: 360 },
    duration: { value: 30 },
  });
  Object.defineProperties(viewport, {
    clientWidth: { value: 640 },
    clientHeight: { value: 360 },
  });
  vi.spyOn(viewport, "getBoundingClientRect").mockReturnValue({
    left: 0,
    top: 0,
    width: 640,
    height: 360,
  } as DOMRect);
  fireEvent.wheel(viewport, { deltaY: -200, clientX: 400, clientY: 180 });
  expect(video.style.transform).not.toContain("scale(1)");
  expect(video.style.transform).not.toContain("translate(0px, 0px)");
  expect(viewport.contains(screen.getByRole("button", { name: "Play" }))).toBe(
    false,
  );
  fireEvent.loadedMetadata(video);
  fireEvent.change(screen.getByRole("slider", { name: "Seek video" }), {
    target: { value: "12" },
  });
  expect(video.currentTime).toBe(12);
  const loopButton = screen.getByRole("button", { name: "Loop video" });
  expect(video.loop).toBe(false);
  fireEvent.click(loopButton);
  expect(video.loop).toBe(true);
  expect(loopButton.getAttribute("aria-pressed")).toBe("true");
  fireEvent.click(loopButton);
  expect(video.loop).toBe(false);
  fireEvent.timeUpdate(video);
  expect(screen.getByLabelText("Playback time").textContent).toBe(
    "0:12 / 0:30",
  );
  fireEvent.click(screen.getByRole("button", { name: "Forward 10 seconds" }));
  expect(video.currentTime).toBe(22);
  fireEvent.click(screen.getByRole("button", { name: "Forward 10 seconds" }));
  expect(video.currentTime).toBe(30);
  fireEvent.keyDown(screen.getByLabelText("Video player"), { key: "j" });
  expect(video.currentTime).toBe(20);
  fireEvent.change(screen.getByRole("combobox", { name: "Playback speed" }), {
    target: { value: "1.5" },
  });
  expect(video.playbackRate).toBe(1.5);
  fireEvent.click(screen.getByRole("button", { name: "Mute" }));
  expect(video.muted).toBe(true);
  fireEvent.click(screen.getByRole("button", { name: "Reset zoom" }));
  expect(video.style.transform).toBe("translate(0px, 0px) scale(1)");
  fireEvent.click(screen.getByRole("button", { name: "Zoom in" }));
  expect(video.style.transform).toContain("scale(1.25)");
  fireEvent.click(screen.getByRole("button", { name: "Zoom out" }));
  expect(video.style.transform).toBe("translate(0px, 0px) scale(1)");
});
beforeEach(() => {
  history.replaceState(null, "", "/");
});
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});
describe("directory previews", () => {
  const directory = {
    ...item,
    name: "album",
    path: "/album",
    type: "dir",
    url: "/?path=%2Falbum%2F",
  };
  it.each([1, 2, 3, 4])(
    "renders a %i-image collage and opens the folder",
    async (count) => {
      const previews = Array.from({ length: count }, (_, i) => `/fs/${i}.png`);
      const fetcher = vi.fn(() =>
        response({ ...collection, items: [{ ...directory, previews }] }),
      );
      vi.stubGlobal("fetch", fetcher);
      render(<App />);
      const link = await screen.findByRole("link", { name: "Open album" });
      expect(link.querySelector(`.directory-collage-${count}`)).not.toBeNull();
      const images = link.querySelectorAll("img");
      expect(images).toHaveLength(count);
      images.forEach((image, i) => {
        expect(image.getAttribute("src")).toBe(previews[i]);
        expect(image.getAttribute("loading")).toBe("lazy");
      });
      expect(link.textContent).toContain("Folder");
      fireEvent.click(images[0]);
      await waitFor(() =>
        expect(fetcher).toHaveBeenLastCalledWith(
          "/api/gallery?path=%2Falbum%2F",
          { signal: expect.any(AbortSignal) },
        ),
      );
      expect(screen.queryByRole("dialog")).toBeNull();
    },
  );
  it("replaces only failed cells with a folder placeholder", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        response({
          ...collection,
          items: [{ ...directory, previews: ["/a.png", "/b.png"] }],
        }),
      ),
    );
    render(<App />);
    const link = await screen.findByRole("link", { name: "Open album" });
    fireEvent.error(link.querySelector("img")!);
    expect(link.querySelectorAll("img")).toHaveLength(1);
    expect(
      link.querySelector(".directory-cell .placeholder")?.textContent,
    ).toBe("▰");
    expect(link.textContent).toContain("Folder");
  });
  it("keeps the folder icon when no preview is available", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => response({ ...collection, items: [directory] })),
    );
    render(<App />);
    const link = await screen.findByRole("link", { name: "Open album" });
    expect(link.querySelector("img")).toBeNull();
    expect(link.querySelector(".placeholder")?.textContent).toBe("▰");
  });
});
describe("routing", () => {
  it("ignores queued callbacks from an old infinite-scroll observer", async () => {
    const callbacks: Array<() => void> = [];
    vi.stubGlobal(
      "IntersectionObserver",
      class {
        constructor(
          callback: (entries: Array<{ isIntersecting: boolean }>) => void,
        ) {
          callbacks.push(() => callback([{ isIntersecting: true }]));
        }
        observe() {}
        disconnect() {}
      },
    );
    const second = { ...item, name: "second.png", path: "/second.png" };
    const fetcher = vi.fn((url: string) =>
      response(
        new URL(url, location.origin).searchParams.get("p") === "2"
          ? { ...collection, page: 2, items: [second] }
          : collection,
      ),
    );
    vi.stubGlobal("fetch", fetcher);
    render(<App />);
    await screen.findByRole("link", { name: "photo.png" });
    const queued = callbacks[0];
    await act(async () => {
      queued();
      // Deliver another notification after the request settles, while React
      // is still batching the page update inside this act scope.
      for (let i = 0; i < 10; i++) await Promise.resolve();
      queued();
    });
    await screen.findByRole("link", { name: "second.png" });
    await act(async () => queued());
    expect(fetcher).toHaveBeenCalledTimes(2);
    expect(screen.getAllByRole("link", { name: "second.png" })).toHaveLength(1);
  });
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
  it("preserves the subdirectory and resets pagination when applying filters", async () => {
    const path = "/albums/c++ & cats/";
    history.replaceState(
      null,
      "",
      "/?" + new URLSearchParams({ path, p: "2" }),
    );
    const fetcher = vi.fn(() => response({ ...collection, page: 2 }));
    vi.stubGlobal("fetch", fetcher);
    render(<App />);
    await screen.findByRole("link", { name: "photo.png" });
    fireEvent.change(screen.getByPlaceholderText("Find a filename"), {
      target: { value: "cat" },
    });
    fireEvent.submit(
      screen.getByPlaceholderText("Find a filename").closest("form")!,
    );
    await waitFor(() => expect(location.search).toContain("q=cat"));
    expect(new URLSearchParams(location.search).has("p")).toBe(false);
    expect(new URLSearchParams(location.search).get("path")).toBe(path);
  });
  it.each(["search", "type", "extension"])(
    "keeps the current directory when a %s filter applies automatically",
    async (filter) => {
      const path = "/albums/c++ & cats/";
      vi.stubGlobal(
        "fetch",
        vi.fn(() => response(collection)),
      );
      render(<App />);
      await screen.findByRole("link", { name: "photo.png" });
      act(() => {
        history.pushState(null, "", "/?" + new URLSearchParams({ path }));
        window.dispatchEvent(new PopStateEvent("popstate"));
      });
      await screen.findByRole("heading", { name: "c++ & cats" });
      if (filter === "type") {
        fireEvent.click(screen.getByLabelText("image"));
      } else {
        fireEvent.input(
          screen.getByPlaceholderText(
            filter === "search" ? "Find a filename" : ".png, .mp4",
          ),
          { target: { value: filter === "search" ? "cat" : ".png" } },
        );
      }
      const key = { search: "q", type: "type", extension: "ext" }[filter]!;
      await waitFor(() =>
        expect(new URLSearchParams(location.search).has(key)).toBe(true),
      );
      expect(new URLSearchParams(location.search).get("path")).toBe(path);
      fireEvent.click(screen.getByRole("button", { name: "Clear" }));
      expect(new URLSearchParams(location.search).get("path")).toBe(path);
      expect(new URLSearchParams(location.search).has(key)).toBe(false);
    },
  );
});
describe("viewer", () => {
  it("allows video picture swipes while reserving native playback controls", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => response({ ...media, item: { ...item, type: "video" } })),
    );
    const navigate = vi.fn();
    render(
      <Viewer
        url="/post/?path=x"
        overlay
        close={() => {}}
        navigate={navigate}
      />,
    );
    await screen.findByRole("heading", { name: "photo.png" });
    const video = document.querySelector("video")!;
    vi.spyOn(video, "getBoundingClientRect").mockReturnValue({
      bottom: 500,
    } as DOMRect);
    const swipe = (y: number) => {
      fireEvent.touchStart(video, {
        touches: [{ identifier: 1, clientX: 200, clientY: y }],
      });
      fireEvent.touchEnd(video, {
        touches: [],
        changedTouches: [{ identifier: 1, clientX: 80, clientY: y }],
      });
    };
    swipe(200);
    expect(navigate).toHaveBeenCalledWith(media.next);
    navigate.mockClear();
    swipe(460);
    expect(navigate).not.toHaveBeenCalled();
  });
  it("navigates deliberate swipes but ignores scrolling, multitouch, cancellation, and boundaries", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => response(media)),
    );
    const navigate = vi.fn();
    render(
      <Viewer
        url="/post/?path=x"
        overlay
        close={() => {}}
        navigate={navigate}
      />,
    );
    const image = await screen.findByRole("img");
    const touch = (x: number, y = 100) => ({
      identifier: 1,
      clientX: x,
      clientY: y,
    });
    const start = () => fireEvent.touchStart(image, { touches: [touch(200)] });
    const end = (x: number, y = 100) =>
      fireEvent.touchEnd(image, { touches: [], changedTouches: [touch(x, y)] });
    start();
    end(80);
    expect(navigate).toHaveBeenCalledWith(media.next);
    navigate.mockClear();
    start();
    end(320); // No previous item.
    start();
    end(180); // A tap or small movement.
    start();
    fireEvent.touchMove(image, { touches: [touch(150, 160)] });
    end(80);
    start();
    fireEvent.touchMove(image, {
      touches: [touch(150), { ...touch(160), identifier: 2 }],
    });
    end(80);
    start();
    fireEvent.touchCancel(image);
    end(80);
    expect(navigate).not.toHaveBeenCalled();
  });
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
  it.each([
    undefined,
    "/vidthumb/?path=/clip.mp4",
    "/vidthumb/?path=/clip.mp4&variant=sheet",
  ])(
    "uses video poster %s and releases sources on unmount",
    async (thumbnail) => {
      vi.stubGlobal(
        "fetch",
        vi.fn(() =>
          response({ ...media, item: { ...item, type: "video", thumbnail } }),
        ),
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
      expect(video.getAttribute("poster")).toBe(thumbnail ?? null);
      expect(video.autoplay).toBe(false);
      view.unmount();
      expect(video.pause).toHaveBeenCalled();
      expect(video.hasAttribute("src")).toBe(false);
      expect(video.load).toHaveBeenCalled();
    },
  );
});
