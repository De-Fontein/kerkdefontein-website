// Moves through the enlarged flyers. Vorige/Volgende work natively (they open the neighbour on top of the
// current flyer); this script closes the current one first, adds arrow keys, swipes and zoom, and preloads.
type Direction = "prev" | "next";

function openFlyer(): HTMLElement | null {
  return document.querySelector<HTMLElement>(".flyer-popover:popover-open");
}

function stepButton(flyer: HTMLElement, direction: Direction): HTMLButtonElement | null {
  return flyer.querySelector<HTMLButtonElement>(`[data-flyer-${direction}]`);
}

function neighbour(from: HTMLElement, direction: Direction): HTMLElement | null {
  const id = stepButton(from, direction)?.getAttribute("popovertarget");
  return id ? document.getElementById(id) : null;
}

// Focus moves to the same button in the next flyer, so repeated presses keep going in that direction.
function go(from: HTMLElement, direction: Direction): void {
  const to = neighbour(from, direction);
  if (!to || to === from) return;
  from.hidePopover();
  to.showPopover();
  stepButton(to, direction)?.focus();
}

function onNavButton(event: MouseEvent): void {
  const button = (event.target as Element).closest("[data-flyer-prev], [data-flyer-next]");
  const from = button?.closest<HTMLElement>(".flyer-popover");
  if (!button || !from) return;
  event.preventDefault(); // replaces the native "open on top"
  go(from, button.hasAttribute("data-flyer-prev") ? "prev" : "next");
}

function onArrowKey(event: KeyboardEvent): void {
  const from = openFlyer();
  if (!from || (event.key !== "ArrowLeft" && event.key !== "ArrowRight")) return;
  event.preventDefault();
  go(from, event.key === "ArrowLeft" ? "prev" : "next");
}

type Point = { x: number; y: number };

// A horizontal touch swipe of at least 50 px; mostly vertical movement is left to scrolling.
function swipeDirection(start: Point, end: PointerEvent): Direction | null {
  const dx = end.clientX - start.x;
  const horizontal = Math.abs(dx) >= 50 && Math.abs(dx) >= Math.abs(end.clientY - start.y);
  if (end.pointerType === "mouse" || !horizontal) return null;
  return dx < 0 ? "next" : "prev";
}

function onSwipe(): void {
  let start: Point = { x: 0, y: 0 };
  document.addEventListener("pointerdown", (event) => {
    start = { x: event.clientX, y: event.clientY };
  });
  document.addEventListener("pointerup", (event) => {
    const from = openFlyer();
    const direction = swipeDirection(start, event);
    if (from && direction) go(from, direction);
  });
}

// An opened flyer takes focus, unless stepping already put it on a button, and tells screen readers it is modal.
function takeFocus(flyer: HTMLElement): void {
  flyer.setAttribute("aria-modal", "true");
  if (!flyer.contains(document.activeElement)) flyer.querySelector<HTMLElement>(".flyer-close")?.focus();
}

// The neighbours load as soon as a flyer opens, so stepping to them shows the image at once.
function preloadNeighbours(flyer: HTMLElement): void {
  for (const direction of ["prev", "next"] as const) {
    neighbour(flyer, direction)?.querySelector("img")?.setAttribute("loading", "eager");
  }
}

// A closed flyer forgets its zoom and hands focus to its own thumbnail, unless stepping opened the next flyer.
// The browser would return focus to the thumbnail that opened the first flyer instead.
function onClose(flyer: HTMLElement): void {
  flyer.classList.remove("zoomed");
  if (!openFlyer()) document.querySelector<HTMLElement>(`.flyer-open[popovertarget="${flyer.id}"]`)?.focus();
}

function onToggle(event: Event): void {
  const flyer = event.target as HTMLElement;
  if (!flyer.matches(".flyer-popover")) return;
  if (!flyer.matches(":popover-open")) return onClose(flyer);
  takeFocus(flyer);
  preloadNeighbours(flyer);
}

// Tab and Shift+Tab cycle through the open flyer's own controls: it covers the page, so focus behind it would
// be invisible (WCAG 2.2 Focus Not Obscured).
function trapTab(event: KeyboardEvent): void {
  const flyer = openFlyer();
  if (!flyer || event.key !== "Tab") return;
  const controls = [...flyer.querySelectorAll<HTMLElement>("a[href], button")];
  const step = event.shiftKey ? -1 : 1;
  const current = controls.indexOf(document.activeElement as HTMLElement);
  event.preventDefault();
  controls.at((current + step) % controls.length)?.focus();
}

// A click on the flyer switches between fitting the screen and a wide, scrollable view. Without this script
// it is a plain link to the image file.
function onZoom(event: MouseEvent): void {
  const flyer = (event.target as Element).closest(".flyer-zoom")?.closest<HTMLElement>(".flyer-popover");
  if (!flyer) return;
  event.preventDefault();
  flyer.classList.toggle("zoomed");
}

// The flyer covers the whole screen, so the browser's "click outside closes it" never fires; a click on its
// empty space does the same.
function onEmptySpace(event: MouseEvent): void {
  const target = event.target as HTMLElement;
  if (target.matches(".flyer-popover:popover-open")) target.hidePopover();
}

document.addEventListener("toggle", onToggle, true); // toggle does not bubble
document.addEventListener("click", onZoom);
document.addEventListener("click", onEmptySpace);
document.addEventListener("click", onNavButton);
document.addEventListener("keydown", onArrowKey);
document.addEventListener("keydown", trapTab);
onSwipe();
