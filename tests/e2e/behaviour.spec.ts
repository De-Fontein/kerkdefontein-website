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

// A slow connection paints a page before all of it has arrived (user, 2026-10-03: shifting is the most annoying
// thing on websites). Simulate every cut-off point: whatever is already drawn must not move when more arrives.
const allPages = ["/", "/eerste-keer/", "/samenkomsten/", "/activiteiten/", "/agenda/", "/over-ons/", "/over-ons/israel/",
  "/over-ons/baptisme/", "/doneren/", "/verhuur/", "/privacy/", "/bestaat-niet"];
for (const path of allPages) {
  test(`${path} never moves what is drawn while the rest arrives`, async ({ page }) => {
    for (const width of [320, 390, 820, 1280]) {
      await page.setViewportSize({ width, height: 800 });
      await page.goto(path);
      const moved = await page.evaluate(() => {
        const visible = (e: Element) => { const r = e.getBoundingClientRect(); return r.width > 0 && r.height > 0; };
        // Inline text rewrapping as more words arrive is how text renders; boxes (blocks, buttons, images) must stay.
        const isBox = (e: Element) => getComputedStyle(e).display !== "inline";
        // Elements and text, in document order: a page that has half arrived has both, up to the same point.
        const all: Node[] = [];
        const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_ELEMENT | NodeFilter.SHOW_TEXT);
        while (walker.nextNode()) all.push(walker.currentNode);
        const elements = all.filter((n): n is Element => n instanceof Element);
        const before = new Map(elements.map((e) => [e, e.getBoundingClientRect()]));
        const problems: string[] = [];
        for (let cut = all.length - 1; cut > 0; cut--) {
          if (!all[cut].isConnected) continue;
          const missing = all[cut];
          // Chromium holds the first paint until <main> starts (rel=expect); the header has its own test below.
          if (!(document.querySelector("main")!.compareDocumentPosition(missing) & Node.DOCUMENT_POSITION_CONTAINED_BY)
            && !(document.querySelector("main")!.compareDocumentPosition(missing) & Node.DOCUMENT_POSITION_FOLLOWING)) continue;
          (missing as ChildNode).remove(); // removing from the end, in document order, rewinds to an earlier cut-off
          for (const e of elements) {
            if (!e.isConnected || !visible(e) || !isBox(e) || !(e.compareDocumentPosition(missing) & Node.DOCUMENT_POSITION_FOLLOWING)) continue;
            const was = before.get(e)!;
            const now = e.getBoundingClientRect();
            if (Math.abs(now.x - was.x) > 0.5 || Math.abs(now.y - was.y) > 0.5) {
              const what = missing instanceof Element ? `<${missing.nodeName.toLowerCase()} class="${missing.className}">` : `text "${missing.textContent!.trim().slice(0, 30)}"`;
              problems.push(`<${e.nodeName.toLowerCase()} class="${e.className}"> moves when ${what} is missing`);
            }
          }
          if (problems.length) break;
        }
        return problems.slice(0, 3);
      });
      expect(moved, `at ${width}px`).toEqual([]);
    }
  });
}

test("every image reserves its space before it loads", async ({ page }) => {
  for (const path of allPages) {
    await page.goto(path);
    const unsized = await page.locator("img:not([width][height])").evaluateAll((imgs) => imgs.map((i) => (i as HTMLImageElement).src));
    expect(unsized, path).toEqual([]);
  }
});
