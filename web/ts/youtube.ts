// YouTube is contacted only after a click, so a page view sends nothing to Google (no cookie banner needed).
for (const button of document.querySelectorAll<HTMLButtonElement>("button[data-youtube-id]")) {
  button.addEventListener("click", () => {
    const id = button.dataset.youtubeId;
    if (!id) return;
    const frame = document.createElement("iframe");
    frame.src = `https://www.youtube-nocookie.com/embed/${encodeURIComponent(id)}?autoplay=1`;
    frame.title = button.dataset.youtubeTitle ?? "YouTube-video";
    frame.allow = "autoplay; encrypted-media; picture-in-picture; fullscreen";
    frame.allowFullscreen = true;
    button.replaceWith(frame);
    frame.focus();
  });
}
