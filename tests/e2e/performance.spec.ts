import { test, expect } from "@playwright/test";

// Lighthouse's mobile profile: slow 4G (150 ms RTT, 1.6 Mbps down, 750 kbps up) and a 4x slower CPU.
const SLOW_4G = {
  offline: false,
  latency: 150,
  downloadThroughput: (1638.4 * 1024) / 8,
  uploadThroughput: (750 * 1024) / 8,
};

type Metrics = { lcp: number; cls: number; bytes: number; script: number };

for (const path of ["/", "/agenda/", "/doneren/"]) {
  test(`${path} stays within the performance budget on slow 4G`, async ({ page, context }, testInfo) => {
    test.skip(testInfo.project.name !== "mobile-chromium", "CDP throttling is Chromium-only; one mobile run is enough");
    const cdp = await context.newCDPSession(page);
    await cdp.send("Network.enable");
    await cdp.send("Network.emulateNetworkConditions", SLOW_4G);
    await cdp.send("Emulation.setCPUThrottlingRate", { rate: 4 });

    await page.goto(path, { waitUntil: "load" });
    const m = await page.evaluate(() => new Promise<Metrics>((resolve) => {
      let lcp = 0;
      let cls = 0;
      new PerformanceObserver((list) => { for (const e of list.getEntries()) lcp = e.startTime; })
        .observe({ type: "largest-contentful-paint", buffered: true });
      new PerformanceObserver((list) => {
        for (const e of list.getEntries() as unknown as { hadRecentInput: boolean; value: number }[]) {
          if (!e.hadRecentInput) cls += e.value;
        }
      }).observe({ type: "layout-shift", buffered: true });
      setTimeout(() => {
        const nav = performance.getEntriesByType("navigation")[0] as PerformanceNavigationTiming;
        const resources = performance.getEntriesByType("resource") as PerformanceResourceTiming[];
        // Flyers are excluded from the page-weight budget (spec §3); the video thumbnail (yt-*) is not a flyer.
        const counted = resources.filter((r) => !/\/media\/(?!yt-)/.test(new URL(r.name).pathname));
        resolve({
          lcp,
          cls,
          bytes: nav.transferSize + counted.reduce((sum, r) => sum + r.transferSize, 0),
          script: resources.filter((r) => r.initiatorType === "script").reduce((sum, r) => sum + r.transferSize, 0),
        });
      }, 500);
    }));
    console.log(`${path} ${JSON.stringify(m)}`);
    expect(m.lcp).toBeLessThanOrEqual(1500);
    expect(m.cls).toBeLessThan(0.1);
    expect(m.bytes).toBeLessThanOrEqual(300 * 1024);
    expect(m.script).toBeLessThanOrEqual(10 * 1024);
  });
}
