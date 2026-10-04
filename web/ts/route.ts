// "Route plannen" opens the visitor's own maps app: Apple Maps on Apple devices, and on Android the geo: link,
// which Android hands to the default maps app (or asks which one). Anything else keeps the OpenStreetMap link.
function mapsLink(address: string): string | null {
  const query = encodeURIComponent(address);
  if (/iPhone|iPad|Macintosh/.test(navigator.userAgent)) return `https://maps.apple.com/?daddr=${query}`;
  if (/Android/.test(navigator.userAgent)) return `geo:0,0?q=${query}`;
  return null;
}

for (const link of document.querySelectorAll<HTMLAnchorElement>("a[data-route]")) {
  const href = mapsLink(link.dataset.route ?? "");
  if (href) link.href = href;
}
