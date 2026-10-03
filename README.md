# kerkdefontein.nl

Website of Baptistengemeente De Fontein, Middelburg. A Go program regenerates the static site every minute
from Google Drive (flyers, ANBI documents), Google Calendar (agenda) and the YouTube channel feed.

## Develop

    brew install vips poppler brotli   # or: apt install libvips-tools poppler-utils brotli
    npm ci
    make preview                       # http://localhost:8080, fake data, no Google account needed
    make test                          # Go tests
    make e2e                           # accessibility + behaviour tests in real browsers

## Licence

Code: MIT (see LICENSE). Texts, photos and logo: © Baptistengemeente De Fontein, all rights reserved.
