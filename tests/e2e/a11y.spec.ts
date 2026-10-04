import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

const paths = ["/", "/eerste-keer/", "/samenkomsten/", "/activiteiten/", "/agenda/", "/over-ons/",
  "/over-ons/israel/", "/over-ons/baptisme/", "/doneren/", "/verhuur/", "/privacy/", "/bestaat-niet"];

for (const scheme of ["light", "dark"] as const) {
  for (const path of paths) {
    test(`${path} meets WCAG 2.2 AA in ${scheme} mode`, async ({ page }) => {
      await page.emulateMedia({ colorScheme: scheme });
      await page.goto(path);
      const results = await new AxeBuilder({ page })
        .withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"])
        .analyze();
      expect(results.violations).toEqual([]);
    });
  }
}

test("an open flyer popover is accessible", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("button", { name: /^Vergroot:/ }).first().click();
  const results = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa", "wcag22aa"]).analyze();
  expect(results.violations).toEqual([]);
});

test("the opened Scipio giving form is accessible", async ({ page }) => {
  await page.route("https://referral.socie.nl/**", (route) => route.fulfill({
    contentType: "text/html", body: `<!doctype html><html lang="nl"><title>Scipio</title><main><button>Give</button></main></html>` }));
  await page.goto("/doneren/");
  await page.getByRole("link", { name: "Geef online" }).click();
  await expect(page.locator("iframe")).toBeVisible();
  const results = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa", "wcag22aa"]).analyze();
  expect(results.violations).toEqual([]);
});
