import React from "react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { Player } from "./Player";
import type { Item } from "../api";

beforeEach(() => {
  Object.defineProperties(document, {
    fullscreenEnabled: { configurable: true, get: () => false },
    fullscreenElement: { configurable: true, get: () => null },
    exitFullscreen: { configurable: true, writable: true, value: vi.fn() },
  });
});

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  for (const key of [
    "fullscreenEnabled",
    "fullscreenElement",
    "exitFullscreen",
  ]) {
    Reflect.deleteProperty(document, key);
  }
});

function setup(enabled = false) {
  vi.spyOn(document, "fullscreenEnabled", "get").mockReturnValue(enabled);
  render(
    <Player
      item={{ type: "video", url: "/fs/test.mp4", name: "test.mp4" } as Item}
    />,
  );
  const video = document.querySelector("video")!;
  const container = video.parentElement!.parentElement!;
  const enter = vi.fn();
  const exit = vi.fn();
  Object.assign(video, {
    webkitEnterFullscreen: enter,
    webkitExitFullscreen: exit,
  });
  return { video, container, enter, exit };
}

it("uses standard container fullscreen and updates the exit control", async () => {
  const { container, enter } = setup(true);
  const request = vi.fn().mockResolvedValue(undefined);
  Object.assign(container, { requestFullscreen: request });
  const exit = vi
    .spyOn(document, "exitFullscreen")
    .mockResolvedValue(undefined);
  fireEvent.click(screen.getByRole("button", { name: "Fullscreen" }));
  expect(request).toHaveBeenCalledOnce();
  expect(enter).not.toHaveBeenCalled();
  const element = vi
    .spyOn(document, "fullscreenElement", "get")
    .mockReturnValue(container);
  fireEvent(document, new Event("fullscreenchange"));
  fireEvent.click(screen.getByRole("button", { name: "Exit fullscreen" }));
  expect(exit).toHaveBeenCalledOnce();
  element.mockReturnValue(null);
  fireEvent(document, new Event("fullscreenchange"));
  expect(screen.getByRole("button", { name: "Fullscreen" })).toBeTruthy();
});

it.each([false, true])(
  "enters native fullscreen synchronously when standard fullscreen is disabled or absent (%s)",
  (enabled) => {
    const { video, container, enter, exit } = setup(enabled);
    const request = vi.fn();
    Object.assign(container, {
      requestFullscreen: enabled ? undefined : request,
    });
    fireEvent.click(screen.getByRole("button", { name: "Fullscreen" }));
    expect(enter).toHaveBeenCalledOnce();
    expect(request).not.toHaveBeenCalled();
    fireEvent(video, new Event("webkitbeginfullscreen"));
    fireEvent.click(screen.getByRole("button", { name: "Exit fullscreen" }));
    expect(exit).toHaveBeenCalledOnce();
    fireEvent(video, new Event("webkitendfullscreen"));
    fireEvent.click(screen.getByRole("button", { name: "Fullscreen" }));
    expect(enter).toHaveBeenCalledTimes(2);
  },
);

it.each(["reject", "throw"])(
  "falls back when standard fullscreen requests %s",
  async (failure) => {
    const { container, enter } = setup(true);
    Object.assign(container, {
      requestFullscreen: vi.fn(() => {
        if (failure === "throw") throw new Error("Unavailable");
        return Promise.reject(new Error("Unavailable"));
      }),
    });
    await act(async () =>
      fireEvent.click(screen.getByRole("button", { name: "Fullscreen" })),
    );
    expect(enter).toHaveBeenCalledOnce();
  },
);

it("handles unavailable and throwing native methods without changing playback", () => {
  const { video, enter, exit } = setup();
  enter.mockImplementation(() => {
    throw new Error("Unavailable");
  });
  fireEvent.click(screen.getByRole("button", { name: "Fullscreen" }));
  Object.assign(video, { webkitEnterFullscreen: undefined });
  fireEvent.click(screen.getByRole("button", { name: "Fullscreen" }));
  Object.assign(video, { webkitDisplayingFullscreen: true });
  exit.mockImplementation(() => {
    throw new Error("Unavailable");
  });
  fireEvent.click(screen.getByRole("button", { name: "Fullscreen" }));
  expect(exit).toHaveBeenCalledOnce();
  expect(video.paused).toBe(true);
  expect(video.getAttribute("src")).toBe("/fs/test.mp4");
});

it("removes standard and native fullscreen event listeners on unmount", () => {
  const documentAdd = vi.spyOn(document, "addEventListener");
  const videoAdd = vi.spyOn(HTMLVideoElement.prototype, "addEventListener");
  const { video } = setup();
  const documentRemove = vi.spyOn(document, "removeEventListener");
  const videoRemove = vi.spyOn(video, "removeEventListener");
  cleanup();
  for (const [name, handler] of documentAdd.mock.calls.filter(
    ([name]) => name === "fullscreenchange",
  )) {
    expect(documentRemove).toHaveBeenCalledWith(name, handler);
  }
  for (const [name, handler] of videoAdd.mock.calls.filter(
    ([name]) =>
      name === "webkitbeginfullscreen" || name === "webkitendfullscreen",
  )) {
    expect(videoRemove).toHaveBeenCalledWith(name, handler);
  }
  expect(
    videoRemove.mock.calls.filter(([name]) => name.startsWith("webkit")),
  ).toHaveLength(2);
});
