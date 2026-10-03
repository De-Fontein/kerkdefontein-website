import { test, expect } from "@playwright/test";

test("a page load makes no third-party requests and no CSP violations", async ({ page, baseURL }) => {
  const foreign: string[] = [];
  const cspErrors: string[] = [];
  page.on("request", (r) => { if (!r.url().startsWith(baseURL!)) foreign.push(r.url()); });
  page.on("console", (m) => { if (m.type() === "error" && /Content.Security.Policy/i.test(m.text())) cspErrors.push(m.text()); });
  for (const path of ["/", "/agenda/", "/doneren/"]) {
    await page.goto(path);
    await page.waitForLoadState("networkidle");
  }
  expect(foreign).toEqual([]);
  expect(cspErrors).toEqual([]);
});

test("the large flyer loads only when its popover opens", async ({ page }) => {
  const large: string[] = [];
  page.on("request", (r) => { if (r.url().includes("-1600.webp")) large.push(r.url()); });
  await page.goto("/");
  await page.waitForLoadState("networkidle");
  expect(large).toEqual([]);
  await page.getByRole("button", { name: /^Vergroot:/ }).first().click();
  await expect(page.locator(".flyer-popover:popover-open img")).toBeVisible();
  // Chromium starts the lazy load a moment after the popover becomes visible.
  await expect.poll(() => large.length).toBe(1);
});

test("a flyer with a link shows Meer info inside the popover", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("button", { name: "Vergroot: Aanbiddingsavond vr 19.30" }).click();
  await expect(page.getByRole("link", { name: "Meer info" })).toHaveAttribute("href", "https://kerkdefontein.nl/agenda/");
});

test("clicking the video loads the youtube-nocookie player", async ({ page }) => {
  await page.route("https://www.youtube-nocookie.com/**", (route) => route.abort());
  await page.goto("/");
  await page.getByRole("button", { name: /^Afspelen:/ }).click();
  await expect(page.locator("iframe")).toHaveAttribute("src", /^https:\/\/www\.youtube-nocookie\.com\/embed\/sample1/);
});

test("the mobile menu opens without JavaScript", async ({ browser, baseURL }) => {
  const context = await browser.newContext({ javaScriptEnabled: false, viewport: { width: 390, height: 844 } });
  const page = await context.newPage();
  await page.goto(baseURL + "/");
  await page.getByRole("button", { name: "Menu" }).click();
  await expect(page.getByRole("link", { name: "Agenda", exact: true })).toBeVisible();
  await context.close();
});

test("unknown paths return the 404 page", async ({ page }) => {
  const response = await page.goto("/bestaat-niet");
  expect(response?.status()).toBe(404);
  await expect(page.getByRole("heading", { name: "Pagina niet gevonden" })).toBeVisible();
});

test("keyboard focus is visible on the video button and high-contrast in light mode", async ({ page }) => {
  await page.emulateMedia({ colorScheme: "light" });
  await page.goto("/");
  const button = page.getByRole("button", { name: /^Afspelen:/ });
  await button.focus();
  const ring = await button.evaluate((el) => {
    const style = getComputedStyle(el);
    return {
      focusVisible: el.matches(":focus-visible"),
      color: style.outlineColor,
      width: parseFloat(style.outlineWidth),
      clippedByParent: getComputedStyle(el.parentElement!).overflow !== "visible",
    };
  });
  expect(ring.focusVisible).toBe(true);
  expect(ring.width).toBeGreaterThanOrEqual(2);
  expect(ring.clippedByParent).toBe(false); // an overflow:hidden wrapper cut the ring off entirely
  expect(ring.color).toBe("rgb(5, 93, 117)"); // teal is 7.4:1 on white; the old sky ring was 2.6:1 (needs 3:1)
});

// Scipio's form is stubbed: CI never contacts Scipio, and a real click would create a pending donation.
const scipio = "https://referral.socie.nl/**";
function stubScipio(body: string) {
  return (route: import("@playwright/test").Route) =>
    route.fulfill({ contentType: "text/html", body: `<!doctype html><html lang="nl"><title>Scipio</title><main>${body}</main></html>` });
}

test("the Scipio giving form loads only after Geef online is clicked", async ({ page }) => {
  const requests: string[] = [];
  page.on("request", (r) => { if (r.url().includes("socie.nl")) requests.push(r.url()); });
  await page.route(scipio, stubScipio("<button>Give</button>"));
  await page.goto("/doneren/");
  await page.waitForLoadState("networkidle");
  expect(requests).toEqual([]);
  await page.getByRole("link", { name: "Geef online" }).click();
  const frame = page.locator("iframe[title='Online geven via Scipio']");
  await expect(frame).toHaveAttribute("src", /^https:\/\/referral\.socie\.nl\/collections\/HC9DTZ7CKV\?/);
  await expect(page).toHaveURL(/\/doneren\/$/);
  await expect(page.frameLocator("iframe").getByRole("button", { name: "Give" })).toBeVisible();
});

// Scipio opens checkout with window.open(url, "_top") after an API call, so the click's activation may have expired.
test("the Scipio form can send the whole page to checkout after a slow response", async ({ page, baseURL }) => {
  await page.route(scipio, stubScipio(
    `<button onclick="setTimeout(() => window.open('${baseURL}/agenda/', '_top'), 6000)">Give</button>`));
  await page.goto("/doneren/");
  await page.getByRole("link", { name: "Geef online" }).click();
  await page.frameLocator("iframe").getByRole("button", { name: "Give" }).click();
  await expect(page).toHaveURL(/\/agenda\/$/, { timeout: 10_000 });
});

test("without JavaScript Geef online links to the Scipio page", async ({ browser, baseURL }) => {
  const context = await browser.newContext({ javaScriptEnabled: false, baseURL });
  const page = await context.newPage();
  await page.goto("/doneren/");
  await expect(page.getByRole("link", { name: "Geef online" })).toHaveAttribute("href", "https://link.socie.nl/r/sci/c/HC9DTZ7CKV");
  await context.close();
});

// On a slow connection the page paints before its scripts finish; nothing may move when they do (user, 2026-10-03).
for (const path of ["/", "/doneren/"]) {
  test(`${path} does not shift when its scripts finish loading late`, async ({ page, browserName }) => {
    test.skip(browserName !== "chromium", "layout-shift entries are Chromium-only");
    await page.route(/\/static\/js\/(youtube|giving)\.js/, async (route) => {
      await new Promise((r) => setTimeout(r, 1500));
      await route.continue();
    });
    await page.addInitScript(() => {
      (window as unknown as { cls: number }).cls = 0;
      new PerformanceObserver((list) => {
        for (const e of list.getEntries() as unknown as { hadRecentInput: boolean; value: number }[]) {
          if (!e.hadRecentInput) (window as unknown as { cls: number }).cls += e.value;
        }
      }).observe({ type: "layout-shift", buffered: true });
    });
    await page.goto(path, { waitUntil: "load" });
    await page.waitForTimeout(300);
    expect(await page.evaluate(() => (window as unknown as { cls: number }).cls)).toBe(0);
  });
}
