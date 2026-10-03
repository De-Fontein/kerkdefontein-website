// Scipio is contacted only after a click: its form loads Google Fonts, so a page view must not load it.
// Without JS the button stays a plain link to Scipio's own page.
for (const link of document.querySelectorAll<HTMLAnchorElement>("a[data-giving-src]")) {
  link.addEventListener("click", (event) => {
    const src = link.dataset.givingSrc;
    if (!src) return;
    event.preventDefault();
    const frame = document.createElement("iframe");
    frame.src = src;
    frame.title = "Online geven via Scipio";
    // Top navigation is needed for checkout; Scipio opens it after an API call, when the click's activation may have expired.
    frame.setAttribute("sandbox", "allow-scripts allow-same-origin allow-forms allow-popups allow-popups-to-escape-sandbox allow-top-navigation");
    link.replaceWith(frame);
    frame.focus();
  });
}
