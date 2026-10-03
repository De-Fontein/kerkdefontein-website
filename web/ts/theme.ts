// Loaded as a blocking script in <head>: a saved theme must be applied before the first paint, or the page
// flashes the device theme first. The choice lives in localStorage only; it never reaches the server.
type Theme = "auto" | "light" | "dark";

const storageKey = "theme";
const cycle: Theme[] = ["auto", "light", "dark"];
const labels: Record<Theme, string> = { auto: "automatisch", light: "licht", dark: "donker" };

function savedTheme(): Theme {
  try {
    const value = localStorage.getItem(storageKey);
    return value === "light" || value === "dark" ? value : "auto";
  } catch {
    return "auto"; // storage blocked (private mode): fall back to the device setting
  }
}

function save(theme: Theme): void {
  try {
    if (theme === "auto") localStorage.removeItem(storageKey);
    else localStorage.setItem(storageKey, theme);
  } catch {
    // Without storage the choice lasts until the next page load, which is acceptable.
  }
}

function apply(theme: Theme): void {
  if (theme === "auto") delete document.documentElement.dataset.theme;
  else document.documentElement.dataset.theme = theme;
}

let current = savedTheme();
apply(current);

document.addEventListener("DOMContentLoaded", () => {
  const toggle = document.querySelector<HTMLButtonElement>("[data-theme-toggle]");
  if (!toggle) return;
  const render = () => {
    const label = `Thema: ${labels[current]}`;
    toggle.setAttribute("aria-label", label);
    toggle.title = label;
    toggle.dataset.state = current; // CSS shows the matching icon
  };
  render();
  toggle.hidden = false;
  toggle.addEventListener("click", () => {
    current = cycle[(cycle.indexOf(current) + 1) % cycle.length];
    save(current);
    apply(current);
    render();
  });
});
