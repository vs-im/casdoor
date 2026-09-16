# Brand assets (build input)

This directory is the **default** value of the `BRAND_ASSETS_DIR` Docker build
argument. Its contents are copied into `web/public/brand/` before the frontend is
built, i.e. they end up served at `/brand/<file>`.

Upstream ships no brand assets here on purpose: the stock defaults of the
`CASDOOR_BRAND_*` variables point at the public Casdoor CDN, so a build without a
brand needs no files at all. A white-label deployment keeps its own logos outside
this repository and passes them in:

```
docker build --build-arg BRAND_ASSETS_DIR=<dir inside the build context> ...
```

with `CASDOOR_BRAND_LOGO_URL=/brand/logo.svg` and friends set at runtime. See
`docs/branding.md`.
