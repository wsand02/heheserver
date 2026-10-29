import { test, expect } from "@playwright/test";
test.beforeEach(async ({ page }) => {
  await page.route("**/api/gallery?*", async (route) => {
    const p = new URL(route.request().url()).searchParams;
    const page = Number(p.get("p") || 1);
    await route.fulfill({
      json: {
        items: Array.from({ length: 12 }, (_, i) => {
          const n = (page - 1) * 12 + i + 1;
          return {
            name: `Image ${n}`,
            path: `/${n}.svg`,
            type: "image",
            url: `/fs/${n}.svg`,
            thumbnail: `/fs/${n}.svg`,
            viewerUrl: `/post/?path=%2F${n}.svg&p=${page}`,
            size: 100,
            modified: "2026-01-01",
            width: 640,
            height: 480,
          };
        }),
        path: "/",
        page,
        pages: 3,
        total: 36,
        infinite: p.get("ext") !== "paged",
        version: "test",
      },
    });
  });
  await page.route("**/api/media?*", async (route) => {
    const p = new URL(route.request().url()).searchParams;
    const n = Number((p.get("path") || "/1.svg").match(/\d+/)?.[0]);
    await route.fulfill({
      json: {
        item: {
          name: `Image ${n}`,
          path: `/${n}.svg`,
          type: "image",
          url: `/fs/${n}.svg`,
          size: 100,
          modified: "2026-01-01",
        },
        position: n,
        total: 36,
        previous: n > 1 ? `/post/?path=%2F${n - 1}.svg` : "",
        next: n < 36 ? `/post/?path=%2F${n + 1}.svg` : "",
        galleryUrl: "/?path=%2F",
      },
    });
  });
  await page.route("**/fs/*.svg", (route) =>
    route.fulfill({
      contentType: "image/svg+xml",
      body: '<svg xmlns="http://www.w3.org/2000/svg" width="640" height="480"><rect width="640" height="480" fill="#537b96"/></svg>',
    }),
  );
});
test("infinite-scroll viewer preserves tiles, scroll, focus, and history", async ({
  page,
}) => {
  await page.goto("/?path=%2F");
  await page
    .getByRole("link", { name: "Image 12", exact: true })
    .scrollIntoViewIfNeeded();
  await expect(
    page.getByRole("link", { name: "Image 13", exact: true }),
  ).toBeVisible();
  const tile = page.getByRole("link", { name: "Image 13", exact: true });
  await expect(tile).toHaveCount(1);
  await tile.scrollIntoViewIfNeeded();
  await tile.focus();
  await expect(page.locator(".grid")).toHaveAttribute("aria-busy", "false");
  const y = await page.evaluate(() => scrollY);
  const count = await page.locator("article").count();
  await tile.click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.getByRole("button", { name: "Next item" }).click();
  await expect(page.getByRole("heading", { name: "Image 14" })).toBeVisible();
  await page.goBack();
  await expect(page.getByRole("dialog")).toHaveCount(0);
  expect(await page.locator("article").count()).toBe(count);
  expect(Math.abs((await page.evaluate(() => scrollY)) - y)).toBeLessThan(5);
  await expect(tile).toBeFocused();
  await page.goForward();
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("dialog")).toHaveCount(0);
});
test("standalone links and pagination", async ({ page }, testInfo) => {
  await page.goto("/post/?path=%2F2.svg");
  await expect(page.getByRole("heading", { name: "Image 2" })).toBeVisible();
  await page.screenshot({ path: testInfo.outputPath("viewer.png") });
  await expect(page.getByRole("dialog")).toHaveCount(0);
  await page.getByRole("button", { name: "← Gallery" }).click();
  await expect(page.getByRole("heading", { name: "Your files" })).toBeVisible();
  await page.goto("/?path=%2F&ext=paged");
  await page.getByRole("button", { name: "Next", exact: true }).click();
  await expect(
    page.getByRole("link", { name: "Image 13", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("link", { name: "Image 1", exact: true }),
  ).toHaveCount(0);
  await page.screenshot({ path: testInfo.outputPath("gallery.png") });
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
});

test("modal keyboard focus and realtime search", async ({ page }) => {
  await page.goto("/?path=%2F");
  const search = page.getByPlaceholder("Find a filename");
  await search.fill("Image");
  await expect(page).toHaveURL(/q=Image/);
  await expect(search).toBeFocused();
  const tile = page.getByRole("link", { name: "Image 1", exact: true });
  await tile.click();
  await expect(page.getByRole("heading", { name: "Image 1" })).toBeVisible();
  await expect(page.getByRole("dialog")).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(page.getByRole("link", { name: "Open original" })).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(page.getByRole("button", { name: "✕ Close" })).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(tile).toBeFocused();
});
