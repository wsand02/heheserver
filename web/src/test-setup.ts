import { vi } from "vitest";
Object.defineProperty(window, "matchMedia", {
  value: () => ({
    matches: false,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  }),
});
Object.defineProperty(window, "scrollTo", { value: vi.fn() });
Object.defineProperty(window, "IntersectionObserver", {
  value: class {
    observe() {}
    disconnect() {}
  },
});
Object.defineProperty(HTMLMediaElement.prototype, "pause", { value: vi.fn() });
Object.defineProperty(HTMLMediaElement.prototype, "load", { value: vi.fn() });
