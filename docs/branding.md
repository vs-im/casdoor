# Branding (white-label)

This fork carries **no** product brand of its own: the tree is stock Casdoor plus
features, and every product name, wordmark, favicon and marketing URL a
white-label deployment replaces is read from `CASDOOR_BRAND_*` environment
variables whose defaults are the upstream Casdoor values.

Two rules keep it that way:

1. A build with no `CASDOOR_BRAND_*` variable is indistinguishable from upstream.
2. `scripts/check-no-brand.sh` fails the build when a brand leaks into the tree.

That is what makes `develop` pushable to the fork's remote and mergeable with
`casdoor/casdoor` master without a scrubbing pass.

## Variables

All of them are read at **runtime**, by the server process, so a rebrand needs no
rebuild — only a restart. Empty or unset means "use the upstream default".

| Variable | Default (upstream) | Where it is used |
|---|---|---|
| `CASDOOR_BRAND_NAME` | `Casdoor` | `<title>` of the login shell; display name of the application seeded by `object/init.go`; TOTP issuer fallback; the `{{brand}}` placeholder in every console locale; console texts (product tour, default email/SMS templates in `ProviderEditPage`, Web3 sign-in domain, footer, Web3Auth typed data) |
| `CASDOOR_BRAND_TAGLINE` | `<name> - sign in` | `<meta name="description">` of the login shell |
| `CASDOOR_BRAND_LOGO_URL` | `https://cdn.casbin.org/img/casdoor-logo_1185x256.png` | logo of the seeded application; default logo of a new application/product in the console; sidebar and login wordmark; product tour cover |
| `CASDOOR_BRAND_LOGO_DARK_URL` | the light logo | wordmark on a dark theme |
| `CASDOOR_BRAND_LOGO_MARK_URL` | the light logo | square mark embedded in the built-in email templates (magic link, verification code, invitation) |
| `CASDOOR_BRAND_FAVICON_URL` | `https://cdn.casbin.org/img/favicon.png` | favicon of the login shell; favicon of the organization seeded by `object/init.go` |
| `CASDOOR_BRAND_WEBSITE_URL` | `https://casdoor.org` | homepage URL of the seeded application; website URL of a new organization |
| `CASDOOR_BRAND_EMAIL_SIGNATURE` | `<name> Team` | signature line of the built-in email templates |
| `CASDOOR_BRAND_TOTP_ISSUER` | `<name>` | issuer an authenticator app shows when the organization has no display name |

Implementation: `conf/brand.go` (the single source), `conf/web_config.go` (the
same values are shipped to the frontend inside the existing `jsonWebConfig`
cookie, so the bundle is brand-free too), `web/src/Conf.ts` (`Conf.BrandName` and
friends — mutable module bindings, read them at render time), `web/src/i18n.ts`
(`interpolation.defaultVariables` substitutes `{{brand}}` in the locale bundles,
no `t()` call site passes anything).

`VITE_BRAND_*` variables exist only for running `vite dev` without a backend; the
image never uses them.

## Assets

Logos and favicons are **not** committed. `web/public/brand/` is filled at image
build time from the `BRAND_ASSETS_DIR` build argument (default:
`web/brand-default`, which is empty), and the files land at `/brand/<file>`:

```
docker build --build-arg BRAND_ASSETS_DIR=web/brand -t casdoor-<deployment>:<tag> .
```

`BRAND_ASSETS_DIR` is a path **inside the Docker build context**, so copy the
deployment's assets into a gitignored directory of the checkout (`web/brand/`)
first. Point the URL variables at them:

```
CASDOOR_BRAND_LOGO_URL=/brand/logo.svg
CASDOOR_BRAND_LOGO_DARK_URL=/brand/logo-dark.svg
CASDOOR_BRAND_LOGO_MARK_URL=/brand/logo-mark.png
CASDOOR_BRAND_FAVICON_URL=/brand/favicon.png
```

A root-relative logo mark is resolved against `originFrontend` (or `origin`)
before it goes into an email, because an email client cannot load a relative URL.

## Running a branded deployment

Keep the values in an env file **outside this repository**, next to the
deployment's compose file, and hand it to the container:

```yaml
services:
  casdoor:
    image: casdoor-<deployment>:<tag>
    env_file:
      - brand.env
```

Nothing else changes: the same image with no `brand.env` is upstream Casdoor.

> A deployment keeps `brand.env`, its `brand-assets/` and its image build script
> in its own infrastructure repository, never here.

## Checks

```
scripts/check-no-brand.sh                     # no brand in the tree (CI gate)
go build ./... && go vet ./...
go test ./conf/... ./object/ -run 'Brand|Totp'
cd web && yarn install && yarn run build
node web/scripts/check-brand-render.mjs "Acme Identity"   # locales render the brand
```

`scripts/check-no-brand.sh` greps the tracked and untracked tree for the brand
patterns (`BRAND_PATTERNS` overrides the list) and fails if any brand asset got
committed. Run it before every push of `develop` and before any upstream PR.

`web/scripts/check-brand-render.mjs` initialises i18next the way `src/i18n.ts`
does and asserts that every `{{brand}}` placeholder in every language renders as
the given brand — the proof that a brand-free tree still produces a branded UI.

The branded shell itself (title, description, favicon) is covered by
`routers/static_filter_brand_test.go`, and the variable defaults by
`conf/brand_test.go`.

## Merging upstream master into develop (instructions for an agent)

This is the order that worked on the last sync (18 upstream commits, one
conflict). `develop` is a published branch, so it is always a merge, never a
rebase.

```
git remote add upstream https://github.com/casdoor/casdoor.git   # once
git fetch upstream master
git merge-base --is-ancestor master upstream/master && git branch -f master upstream/master
git switch develop
git merge upstream/master
```

`master` only ever tracks upstream: it must fast-forward, and nothing of ours may
land on it.

Conflict zones, in the order they usually appear:

| Area | What upstream does | What to keep |
|---|---|---|
| `object/application.go` | guards the seeded application by its upstream name and adds authorization checks around it | take upstream's new checks **and** keep the fork's application name in the guard — this was the only conflict of the last sync |
| `conf/conf.go`, `conf/web_config.go` | adds config items | keep both: upstream's items **and** the `Brand*` fields plus their assignments in `GetWebConfig` |
| `conf/brand.go` | does not exist upstream | ours, as it is |
| `object/init.go` | edits the seeded organization/application | take upstream's structure, then re-apply `conf.GetBrandName/LogoUrl/WebsiteUrl/FaviconUrl` in place of the literals it reintroduces |
| `object/mfa_totp.go` | may touch the issuer fallback | keep `conf.GetBrandTotpIssuer()` |
| `object/magic_link.go`, `controllers/magic_link.go`, `routers/lightweight_auth_filter.go` | fork-only files | ours |
| `routers/static_filter.go` | changes how the shell is served | keep upstream's serving logic, keep the `applyBrandToIndexHtml` call right after the organization-theme block |
| `web/index.html` | changes the shell | the `<title>`, the description and `/favicon.png` must stay **byte-identical to the strings `applyBrandToIndexHtml` looks for** |
| `web/src/locales/*/data.json` | new and changed strings | take upstream's text, then replace the product name in the **values** with `{{brand}}` (keys stay upstream) |
| `web/src/lib/setting.tsx`, `web/src/pages/defaults.ts`, `web/src/lib/tour-config.ts` | new defaults with the upstream product name | route them through `Conf.Brand*` / the `{brand}` placeholder |

After the merge, in this order:

```
scripts/check-no-brand.sh
go build ./... && go vet ./...
go test ./conf/... ./object/ -run 'Brand|Totp'
cd web && yarn install && yarn run typecheck && yarn run lint && yarn run build
node scripts/check-brand-render.mjs "<the deployment brand>"
```

`go vet` reports one pre-existing finding in `storage/casdoor.go` (unkeyed
fields) that comes from upstream; `go test ./object/...` as a whole needs a live
database (`TestDumpToFile`), which is why only the targeted tests are listed.

A new upstream string that names the product is a **finding, not a conflict**:
`check-no-brand.sh` will not catch it (it is the upstream brand, not a downstream
one), so grep the merge diff for the product name and decide whether it needs the
`{{brand}}` placeholder.

Finally, confirm the merge really happened before pushing:

```
git merge-base --is-ancestor upstream/master develop   # upstream is in
git log -1 --format=%P                                 # two parents
git diff upstream/master..develop --stat               # only our delta is left
```
