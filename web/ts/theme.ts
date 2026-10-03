// Loaded as a blocking script in <head>: a saved theme must be applied before the first paint, or the page
// flashes the device theme first. The choice lives in localStorage only; it never reaches the server.
// The menu itself is native (popover + radio group); this script only applies, remembers and positions.
type Theme = "auto" | "light" | "dark" | "oled";

const storageKey = "theme";
const labels: Record<Theme, string> = { auto: "automatisch", light: "licht", dark: "donker", oled: "OLED (zwart)" };

function isTheme(value: string | null): value is Theme {
  return value !== null && value in labels;
}

function savedTheme(): Theme {
  try {
    const value = localStorage.getItem(storageKey);
    return isTheme(value) ? value : "auto";
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

type ThemeControls = { toggle: HTMLButtonElement; menu: HTMLElement };

function render({ toggle, menu }: ThemeControls): void {
  const label = `Thema: ${labels[current]}`;
  toggle.setAttribute("aria-label", label);
  toggle.title = label;
  toggle.dataset.state = current; // CSS shows the matching icon
  const radio = menu.querySelector<HTMLInputElement>(`input[value="${current}"]`);
  if (radio) radio.checked = true;
}

function choose(value: string, controls: ThemeControls): void {
  if (!isTheme(value)) return;
  current = value;
  save(current);
  apply(current);
  render(controls);
}

// Arrow keys preview themes and keep the menu open; only a real click (detail > 0) picks one and closes it.
function closeAfterClick(event: MouseEvent, menu: HTMLElement): void {
  const onOption = (event.target as HTMLElement).closest("label") !== null;
  if (event.detail > 0 && onOption) menu.hidePopover();
}

// Anchors the menu under its button; CSS anchor positioning is not yet supported in every browser.
function placeUnder({ toggle, menu }: ThemeControls): void {
  const button = toggle.getBoundingClientRect();
  menu.style.top = `${button.bottom + 8}px`;
  menu.style.right = `${document.documentElement.clientWidth - button.right}px`;
}

function initMenu(controls: ThemeControls): void {
  const { toggle, menu } = controls;
  render(controls);
  toggle.hidden = false;
  menu.addEventListener("change", (event) => choose((event.target as HTMLInputElement).value, controls));
  menu.addEventListener("click", (event) => closeAfterClick(event, menu));
  menu.addEventListener("beforetoggle", (event) => {
    if ((event as ToggleEvent).newState === "open") placeUnder(controls);
  });
}

document.addEventListener("DOMContentLoaded", () => {
  const toggle = document.querySelector<HTMLButtonElement>("[data-theme-toggle]");
  const menu = document.getElementById("theme-menu");
  if (toggle && menu) initMenu({ toggle, menu });
});
