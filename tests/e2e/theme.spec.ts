import { test, expect, type Page } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

const theme = (page: Page) => page.evaluate(() => document.documentElement.dataset.theme ?? "auto");
const background = (page: Page) => page.evaluate(() => getComputedStyle(document.body).backgroundColor);
const toggle = (page: Page) => page.locator("[data-theme-toggle]");

async function choose(page: Page, label: string) {
  await toggle(page).click();
  await page.getByRole("radio", { name: label }).check();
}

test("choosing a theme applies it, updates the button and closes the menu", async ({ page }) => {
  await page.goto("/");
  await expect(toggle(page)).toHaveAttribute("aria-label", "Thema: automatisch");
  for (const [label, value, name] of [["Licht", "light", "licht"], ["Donker", "dark", "donker"],
    ["OLED (zwart)", "oled", "OLED (zwart)"], ["Automatisch", "auto", "automatisch"]]) {
    await choose(page, label);
    expect(await theme(page)).toBe(value);
    await expect(toggle(page)).toHaveAttribute("aria-label", `Thema: ${name}`);
    await expect(page.locator("#theme-menu")).toBeHidden();
  }
});

test("a chosen theme survives a reload and is checked in the menu", async ({ page }) => {
  await page.emulateMedia({ colorScheme: "light" });
  await page.goto("/");
  await choose(page, "Donker");
  await page.reload();
  expect(await theme(page)).toBe("dark");
  expect(await background(page)).toBe("rgb(15, 15, 16)");
  await toggle(page).click();
  await expect(page.getByRole("radio", { name: "Donker" })).toBeChecked();
});

test("forced light mode wins over a dark device setting", async ({ page }) => {
  await page.emulateMedia({ colorScheme: "dark" });
  await page.goto("/");
  await choose(page, "Licht");
  expect(await background(page)).toBe("rgb(255, 255, 255)");
});

for (const [stored, bg] of [["dark", "rgb(15, 15, 16)"], ["oled", "rgb(0, 0, 0)"]]) {
  test(`forced ${stored} theme on a light device meets WCAG 2.2 AA`, async ({ page }) => {
    await page.emulateMedia({ colorScheme: "light" });
    await page.goto("/");
    await page.evaluate((v) => localStorage.setItem("theme", v), stored);
    await page.reload();
    expect(await background(page)).toBe(bg);
    const results = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"]).analyze();
    expect(results.violations).toEqual([]);
  });
}

test("the open theme menu meets WCAG 2.2 AA", async ({ page }) => {
  await page.goto("/");
  await toggle(page).click();
  await expect(page.locator("#theme-menu")).toBeVisible();
  const results = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"]).analyze();
  expect(results.violations).toEqual([]);
});

test("the theme menu works with the keyboard", async ({ page }) => {
  await page.goto("/");
  await toggle(page).focus();
  await page.keyboard.press("Enter");
  await expect(page.locator("#theme-menu")).toBeVisible();
  await page.getByRole("radio", { name: "Automatisch" }).focus();
  await page.keyboard.press("ArrowDown");
  expect(await theme(page)).toBe("light");
  await expect(page.locator("#theme-menu")).toBeVisible(); // arrow keys preview themes; only Escape or a click closes
  await page.keyboard.press("Escape");
  await expect(page.locator("#theme-menu")).toBeHidden();
});

test("without JavaScript there is no theme button", async ({ browser, baseURL }) => {
  const context = await browser.newContext({ javaScriptEnabled: false });
  const page = await context.newPage();
  await page.goto(baseURL + "/");
  await expect(toggle(page)).toBeHidden();
  await context.close();
});

test("the toggle sits after the links on desktop and left of Menu on phones", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto("/");
  const lastLink = page.getByRole("link", { name: "Verhuur", exact: true });
  expect((await toggle(page).boundingBox())!.x).toBeGreaterThan((await lastLink.boundingBox())!.x);

  await page.setViewportSize({ width: 390, height: 844 });
  const menu = page.getByRole("button", { name: "Menu" });
  expect((await toggle(page).boundingBox())!.x).toBeLessThan((await menu.boundingBox())!.x);
});

test("the theme menu opens below its button and stays on screen", async ({ page }) => {
  for (const width of [390, 1280]) {
    await page.setViewportSize({ width, height: 800 });
    await page.goto("/");
    await toggle(page).click();
    const button = (await toggle(page).boundingBox())!;
    const menu = (await page.locator("#theme-menu").boundingBox())!;
    expect(menu.y).toBeGreaterThanOrEqual(button.y + button.height);
    expect(menu.x + menu.width).toBeLessThanOrEqual(width);
    expect(Math.abs(menu.x + menu.width - (button.x + button.width))).toBeLessThanOrEqual(24);
  }
});

test("the theme icon is centred in its button in every state", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto("/");
  for (const label of ["Licht", "Donker", "OLED (zwart)", "Automatisch"]) {
    await choose(page, label);
    const button = (await toggle(page).boundingBox())!;
    const icon = (await toggle(page).locator("svg:visible").boundingBox())!;
    expect(Math.abs(icon.x + icon.width / 2 - (button.x + button.width / 2))).toBeLessThanOrEqual(0.5);
    expect(Math.abs(icon.y + icon.height / 2 - (button.y + button.height / 2))).toBeLessThanOrEqual(0.5);
  }
});
