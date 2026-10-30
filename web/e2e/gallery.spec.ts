import { test, expect } from "@playwright/test";

for (const theme of ["light", "dark"]) {
  test(`directory collages in ${theme} theme`, async ({ page }, testInfo) => {
    await page.route("**/api/gallery?*", (route) =>
      route.fulfill({
        json: {
          items: Array.from({ length: 5 }, (_, count) => ({
            name: `Folder ${count}`,
            path: `/folder${count}`,
            type: "dir",
            url: `/?path=%2Ffolder${count}%2F`,
            viewerUrl: "",
            size: 0,
            modified: "2026-01-01",
            width: 0,
            height: 0,
            previews: Array.from({ length: count }, (_, i) => `/fs/${i}.svg`),
          })),
          path: "/",
          page: 1,
          pages: 1,
          total: 5,
          infinite: false,
          version: "test",
        },
      }),
    );
    await page.goto("/?path=%2F");
    await page.getByRole("combobox", { name: "Theme" }).selectOption(theme);
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    for (let count = 1; count <= 4; count++) {
      const link = page.getByRole("link", {
        name: `Open Folder ${count}`,
        exact: true,
      });
      await link.scrollIntoViewIfNeeded();
      await expect(link.locator("img")).toHaveCount(count);
      await expect(link.locator(".folder-badge")).toBeVisible();
      const cells = await link.locator(".directory-cell").evaluateAll((cells) =>
        cells.map((cell) => {
          const r = cell.getBoundingClientRect();
          return { x: r.x, y: r.y, w: r.width, h: r.height };
        }),
      );
      const box = (await link.boundingBox())!;
      expect(Math.abs(box.width / box.height - 4 / 3)).toBeLessThan(0.02);
      expect(cells.every((cell) => cell.w > 0 && cell.h > 0)).toBe(true);
      if (count === 2) {
        expect(cells[0].y).toBe(cells[1].y);
        expect(cells[1].x).toBeGreaterThan(cells[0].x);
      }
      if (count === 3) {
        expect(cells[0].h).toBeGreaterThan(cells[1].h * 1.9);
        expect(cells[1].x).toBe(cells[2].x);
      }
      if (count === 4) {
        expect(cells[0].y).toBe(cells[1].y);
        expect(cells[2].y).toBeGreaterThan(cells[0].y);
      }
    }
    await page.evaluate(() => scrollTo(0, 0));
    await page.screenshot({
      path: testInfo.outputPath(`collages-${theme}.png`),
      fullPage: true,
    });
    await page
      .getByRole("link", { name: "Open Folder 4", exact: true })
      .click();
    await expect(page).toHaveURL(/path=%2Ffolder4%2F/);
    await expect(page.getByRole("dialog")).toHaveCount(0);
  });
}

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
