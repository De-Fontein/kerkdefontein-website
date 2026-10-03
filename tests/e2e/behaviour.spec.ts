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

test("large flyers load only once one opens, together with its two neighbours", async ({ page }) => {
  const large: string[] = [];
  page.on("request", (r) => { if (r.url().includes("-1600.webp")) large.push(r.url()); });
  await page.goto("/");
  await page.waitForLoadState("networkidle");
  expect(large).toEqual([]);
  await page.getByRole("button", { name: /^Vergroot:/ }).first().click();
  await expect(page.locator(".flyer-popover:popover-open img")).toBeVisible();
  // Chromium starts the lazy load a moment after the popover becomes visible. 3 = the open flyer + neighbours.
  await expect.poll(() => new Set(large).size).toBe(3);
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

test("Activiteiten highlights the activity being read in the list", async ({ page }) => {
  for (const width of [390, 1280]) {
    await page.setViewportSize({ width, height: 800 });
    await page.goto("/activiteiten/");
    await page.evaluate(() => document.getElementById("kringen")!.scrollIntoView({ behavior: "instant" }));
    await expect(page.locator(".act-nav a[aria-current]").filter({ visible: true })).toHaveText(width < 1248 ? [] : ["Kringen"]);
    if (width < 1248) await expect(page.locator("[data-current]")).toHaveText("Kringen");
  }
});

test("on phones Ga naar opens the grouped list and jumps to the chosen activity", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 800 });
  await page.goto("/activiteiten/");
  await page.getByRole("button", { name: /Ga naar/ }).click();
  const list = page.locator("#act-jump-list");
  await expect(list).toBeVisible();
  await list.getByRole("link", { name: "Pastoraat", exact: true }).click();
  await expect(list).toBeHidden();
  await expect(page.locator("#pastoraat")).toBeInViewport();
  await expect(page.locator("[data-current]")).toHaveText("Pastoraat");
});

test("Naar boven returns to the top of Activiteiten on phone and desktop", async ({ page }) => {
  for (const width of [390, 1280]) {
    await page.setViewportSize({ width, height: 800 });
    await page.goto("/activiteiten/");
    await page.evaluate(() => document.getElementById("pastoraat")!.scrollIntoView({ behavior: "instant" }));
    await page.getByRole("link", { name: /Naar boven/ }).filter({ visible: true }).click();
    await expect.poll(() => page.evaluate(() => scrollY)).toBe(0);
  }
});

// Seen on an iPhone (user, 2026-10-03): at ~250 px the name ran under the theme button and a long video title
// filled the thumbnail and was cut off. 250 px covers small phones and zoomed-in Safari.
test("the header and video fit at every width from 250 px", async ({ page }) => {
  for (const width of [250, 320, 390, 820, 1280]) {
    await page.setViewportSize({ width, height: 800 });
    await page.goto("/");
    const brand = (await page.locator(".brand").boundingBox())!;
    const toggle = (await page.locator(".theme-toggle").boundingBox())!;
    expect(brand.x + brand.width, `brand overlaps the theme button at ${width}px`).toBeLessThanOrEqual(toggle.x);
    expect(brand.height, `brand wraps at ${width}px`).toBeLessThanOrEqual(44);
    const facade = (await page.locator(".video-facade").boundingBox())!;
    const label = (await page.locator(".video-play").boundingBox())!;
    expect(label.y + label.height, `play label is cut off at ${width}px`).toBeLessThanOrEqual(facade.y + facade.height);
    expect(label.height, `play label covers the thumbnail at ${width}px`).toBeLessThanOrEqual(facade.height / 2.5);
    await expect(page.getByText("Voorbeelddienst over een lang onderwerp | Spreker | 04-10-2026", { exact: true })).toBeVisible();
  }
});

// Text pages share one centred reading column (user, 2026-10-03: Agenda, then Activiteiten, sat to the left of
// the others). Activiteiten keeps its list in the margin beside that column. Home has its own wide layout.
test("text pages line up in the same centred column", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  const columns: Record<string, number> = {};
  for (const path of allPages.filter((p) => p !== "/")) {
    await page.goto(path);
    columns[path] = Math.round((await page.locator("main h1").boundingBox())!.x);
  }
  const reference = columns["/over-ons/"];
  expect(Object.entries(columns).filter(([, x]) => x !== reference)).toEqual([]);
});

test("arrow keys move through the enlarged flyers and keep focus in the open one", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("button", { name: "Vergroot: Aanbiddingsavond vr 19.30" }).click();
  const open = page.locator(".flyer-popover:popover-open");
  await expect(open.locator(".flyer-count")).toHaveText("1 van 3");
  await page.keyboard.press("ArrowRight");
  await expect(open.locator(".flyer-count")).toHaveText("2 van 3");
  await expect(open).toHaveCount(1);
  expect(await page.evaluate(() => document.activeElement?.closest(".flyer-popover")?.matches(":popover-open"))).toBe(true);
  await page.keyboard.press("ArrowLeft");
  await page.keyboard.press("ArrowLeft"); // wraps from the first to the last
  await expect(open.locator(".flyer-count")).toHaveText("3 van 3");
});

test("Volgende works without JavaScript", async ({ browser, baseURL }) => {
  const context = await browser.newContext({ javaScriptEnabled: false, baseURL });
  const page = await context.newPage();
  await page.goto("/");
  await page.getByRole("button", { name: "Vergroot: Aanbiddingsavond vr 19.30" }).click();
  await page.locator(".flyer-popover:popover-open").getByRole("button", { name: "Volgende flyer" }).click();
  // Natively the next flyer opens on top of the current one; the top one is the last in the top layer.
  await expect(page.locator(".flyer-popover:popover-open .flyer-count").last()).toHaveText("2 van 3");
  await context.close();
});

test("swiping left on an enlarged flyer shows the next one", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("button", { name: "Vergroot: Aanbiddingsavond vr 19.30" }).click();
  const open = page.locator(".flyer-popover:popover-open");
  await open.evaluate((el) => {
    const swipe = (type: string, x: number) =>
      el.dispatchEvent(new PointerEvent(type, { pointerType: "touch", clientX: x, clientY: 300, bubbles: true }));
    swipe("pointerdown", 300);
    swipe("pointerup", 120);
  });
  await expect(open.locator(".flyer-count")).toHaveText("2 van 3");
});

// Seen by the user (2026-10-03): the enlarged flyer grew when its image loaded, moving the buttons. The popup's
// size may depend on the screen only, never on the image, so stepping through flyers keeps the buttons still.
test("the gallery buttons stay put while images load and between flyers", async ({ page }) => {
  await page.route(/-1600\.webp$/, async (route) => {
    await new Promise((r) => setTimeout(r, 800));
    await route.continue();
  });
  for (const [width, height] of [[390, 844], [1280, 800]]) {
    await page.setViewportSize({ width, height });
    await page.goto("/");
    await page.getByRole("button", { name: "Vergroot: Aanbiddingsavond vr 19.30" }).click();
    const next = () => page.locator(".flyer-popover:popover-open [data-flyer-next]").boundingBox();
    const first = await next();
    await expect(page.locator(".flyer-popover:popover-open img")).toHaveJSProperty("complete", true);
    expect(await next(), `moved while the image loaded at ${width}px`).toEqual(first);
    await page.keyboard.press("ArrowRight");
    expect(await next(), `moved to the next flyer at ${width}px`).toEqual(first);
  }
});

// "Route plannen" opens the visitor's own maps app (user, 2026-10-03); OpenStreetMap stays the fallback.
for (const [device, userAgent, expected] of [
  ["iPhone", "Mozilla/5.0 (iPhone; CPU iPhone OS 26_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148", /^https:\/\/maps\.apple\.com\/\?daddr=Roozenburglaan/],
  ["Android", "Mozilla/5.0 (Linux; Android 16; Pixel 9) AppleWebKit/537.36 Chrome/140.0 Mobile Safari/537.36", /^geo:0,0\?q=Roozenburglaan/],
  ["Windows", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/140.0 Safari/537.36", /^https:\/\/www\.openstreetmap\.org\//],
] as const) {
  test(`Route plannen opens the maps app on ${device}`, async ({ browser, baseURL }) => {
    const context = await browser.newContext({ userAgent, baseURL });
    const page = await context.newPage();
    await page.goto("/eerste-keer/");
    await expect(page.getByRole("link", { name: "Route plannen" })).toHaveAttribute("href", expected);
    await context.close();
  });
}

// Fitting the whole flyer to the screen height left it barely larger than its thumbnail on a laptop (user,
// 2026-10-03). The enlarged flyer fills the popup's width and scrolls; its buttons stay pinned in view.
test("an enlarged flyer is about twice its thumbnail on desktop, with its buttons in view", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto("/");
  const thumb = (await page.locator(".flyer-open img").first().boundingBox())!;
  await page.getByRole("button", { name: "Vergroot: Aanbiddingsavond vr 19.30" }).click();
  const open = page.locator(".flyer-popover:popover-open");
  const large = (await open.locator("img").boundingBox())!;
  expect(large.width).toBeGreaterThanOrEqual(thumb.width * 1.9);
  for (const name of ["Volgende flyer", "Sluiten"]) {
    await expect(open.getByRole("button", { name })).toBeInViewport({ ratio: 1 });
  }
  await open.evaluate((el) => el.scrollTo(0, el.scrollHeight));
  await expect(open.getByRole("button", { name: "Volgende flyer" })).toBeInViewport({ ratio: 1 });
});
