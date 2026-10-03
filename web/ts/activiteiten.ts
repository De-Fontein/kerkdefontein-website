// Highlights the activity being read, in the desktop list and the phone's "Ga naar" bar. Everything else on
// the page (jump list, anchors, Naar boven) is native and works without this script.
const links = [...document.querySelectorAll<HTMLAnchorElement>(".act-nav a")];
const headings = [...document.querySelectorAll<HTMLHeadingElement>(".act-article h2[id]")];
const currentLabel = document.querySelector<HTMLElement>("[data-current]");
const jumpList = document.getElementById("act-jump-list");

function mark(heading: HTMLHeadingElement): void {
  for (const link of links) {
    if (link.hash === `#${heading.id}`) link.setAttribute("aria-current", "location");
    else link.removeAttribute("aria-current");
  }
  if (currentLabel) currentLabel.textContent = heading.textContent;
}

// The activity being read is the last heading that has passed the top third of the screen.
function update(): void {
  const passed = headings.filter((h) => h.getBoundingClientRect().top <= innerHeight / 3);
  const heading = passed.at(-1) ?? headings[0];
  if (heading) mark(heading);
}

// At most one update per frame, however fast the scroll events come.
function onEveryFrameWhileScrolling(work: () => void): void {
  let queued = false;
  addEventListener("scroll", () => {
    if (queued) return;
    queued = true;
    requestAnimationFrame(() => {
      queued = false;
      work();
    });
  }, { passive: true });
}

function closeJumpListOnChoice(list: HTMLElement): void {
  list.addEventListener("click", (event) => {
    if ((event.target as HTMLElement).closest("a")) list.hidePopover();
  });
}

onEveryFrameWhileScrolling(update);
if (jumpList) closeJumpListOnChoice(jumpList);
update();
