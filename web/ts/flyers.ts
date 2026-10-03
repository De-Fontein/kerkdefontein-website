// Moves through the enlarged flyers. Vorige/Volgende work natively (they open the neighbour on top of the
// current flyer); this script closes the current one first, adds arrow keys and swipes, and preloads neighbours.
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

// The neighbours load as soon as a flyer opens, so stepping to them shows the image at once.
function preloadNeighbours(event: Event): void {
  const flyer = event.target as HTMLElement;
  if (!flyer.matches(".flyer-popover:popover-open")) return;
  for (const direction of ["prev", "next"] as const) {
    neighbour(flyer, direction)?.querySelector("img")?.setAttribute("loading", "eager");
  }
}

document.addEventListener("toggle", preloadNeighbours, true); // toggle does not bubble
document.addEventListener("click", onNavButton);
document.addEventListener("keydown", onArrowKey);
onSwipe();
