import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

const theme = (page: import("@playwright/test").Page) =>
  page.evaluate(() => document.documentElement.dataset.theme ?? "auto");

test("the theme toggle cycles auto → light → dark → auto", async ({ page }) => {
  await page.goto("/");
  const toggle = page.locator("[data-theme-toggle]");
  await expect(toggle).toHaveAttribute("aria-label", "Thema: automatisch");
  await toggle.click();
  expect(await theme(page)).toBe("light");
  await expect(toggle).toHaveAttribute("aria-label", "Thema: licht");
  await toggle.click();
  expect(await theme(page)).toBe("dark");
  await expect(toggle).toHaveAttribute("aria-label", "Thema: donker");
  await toggle.click();
  expect(await theme(page)).toBe("auto");
});

test("a chosen theme survives a reload and overrides the device setting", async ({ page }) => {
  await page.emulateMedia({ colorScheme: "light" });
  await page.goto("/");
  await page.locator("[data-theme-toggle]").click(); // light
  await page.locator("[data-theme-toggle]").click(); // dark
  await page.reload();
  expect(await theme(page)).toBe("dark");
  const background = await page.evaluate(() => getComputedStyle(document.body).backgroundColor);
  expect(background).toBe("rgb(15, 15, 16)");
});

test("forced light mode wins over a dark device setting", async ({ page }) => {
  await page.emulateMedia({ colorScheme: "dark" });
  await page.goto("/");
  await page.locator("[data-theme-toggle]").click(); // light
  const background = await page.evaluate(() => getComputedStyle(document.body).backgroundColor);
  expect(background).toBe("rgb(255, 255, 255)");
});

test("forced dark mode on a light device meets WCAG 2.2 AA", async ({ page }) => {
  await page.emulateMedia({ colorScheme: "light" });
  await page.goto("/");
  await page.evaluate(() => localStorage.setItem("theme", "dark"));
  await page.reload();
  expect(await page.evaluate(() => getComputedStyle(document.body).backgroundColor)).toBe("rgb(15, 15, 16)");
  const results = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"]).analyze();
  expect(results.violations).toEqual([]);
});

test("without JavaScript there is no toggle", async ({ browser, baseURL }) => {
  const context = await browser.newContext({ javaScriptEnabled: false });
  const page = await context.newPage();
  await page.goto(baseURL + "/");
  await expect(page.locator("[data-theme-toggle]")).toBeHidden();
  await context.close();
});
