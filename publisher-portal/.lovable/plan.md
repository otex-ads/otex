## Goal

Convert the current PropelAds Advertiser console into a Publisher Portal for a Propeller-Ads-style network, keeping the existing visual design (layout, tokens, cards, sidebar shell, animations, typography) untouched. Rewrite the content, routes, data model, and flows around publishers (sites, zones, ad code, payouts) and wire everything to the REST API at `http://167.233.171.202:8080` with JWT auth in `localStorage`.

## Scope of design preservation

- Keep `src/styles.css` tokens, `SurfaceCard`, `StatusPill`, `PageHeader`, `Modal`, `Skeleton`, `EmptyState`, all shadcn/ui components, framer-motion transitions, dashboard hero style, table styling, and sidebar layout in `src/routes/_app.tsx`.
- Only the sidebar branding label ("PropelAds / Advertiser" → "PropelAds / Publisher"), nav items, and page contents change.

## Routes (file-based, under existing `_app` layout)

```
src/routes/
  _app.tsx                       (update nav + branding text only)
  _app.index.tsx                 (Publisher Dashboard)
  _app.sites.tsx                 (Sites list + add/edit/delete modals)
  _app.sites.$siteId.tsx         (Site details: stats, zones, trends)
  _app.zones.tsx                 (Zones list + add/edit/delete modals)
  _app.zones.$zoneId.tsx         (Zone details + ad code)
  _app.adcode.tsx                (Ad Code Generator)
  _app.stats.tsx                 (Statistics/Analytics — rewrite existing)
  _app.reports.tsx               (Reports)
  _app.payouts.tsx               (Earnings, history, request payout, methods)
  _app.settings.tsx              (Profile, password, notifications, API keys, 2FA)
  auth.login.tsx                 (Public login page)
  _app.tsx  → gate: redirect to /auth/login if no JWT
```

Delete advertiser-only routes: `_app.campaigns.tsx`, `_app.campaigns.new.tsx`, `_app.creatives.tsx`, `_app.wallet.tsx` (replaced by payouts).

Sidebar menu becomes: Dashboard, Sites, Zones, Ad Code, Statistics, Reports, Payouts, Settings. Wallet balance widget → "Available balance" widget linking to /payouts.

## Data layer

New API client `src/lib/api.ts`:

- `API_BASE = "http://167.233.171.202:8080"`
- `apiFetch(path, opts)` — attaches `Authorization: Bearer <token>` from `localStorage.getItem("pub_token")`, JSON-encodes body, throws typed errors on non-2xx, 401 → clears token + navigates to `/auth/login`.
- Typed helpers for sites, zones, stats, payouts, balance.

New Zod schemas in `src/lib/schemas.ts` for Site, Zone, Stats, Payout, Balance and all form inputs (add-site, add-zone, request-payout, profile, password).

Auth state:

- `src/lib/auth.ts` — `getToken()`, `setToken()`, `clearToken()`, `useAuth()` hook using a small zustand store (already have `src/lib/store.ts` pattern).
- `_app.tsx` `beforeLoad` (or a client-side effect since routes are SPA-ish here) redirects to `/auth/login` when no token.

TanStack Query used for all reads/mutations; toast notifications via existing `sonner` setup.

## Feature detail (mapped to sections)

1. **Dashboard** — three revenue KPI cards (total, today, this month) reusing `StatCard`; secondary row of performance metrics (impr, clicks, CTR, eCPM, fill rate); revenue line chart with 7/30/90 tabs (recharts, existing color tokens); top sites + top zones tables side-by-side; quick actions row (Add site, Generate code, Request payout) styled like existing hero buttons.
2. **Sites** — table with existing table styles; `Modal` for add/edit forms (react-hook-form + zod); status pill for active/paused; AlertDialog for delete confirmation; row click → site details route.
3. **Zones** — same table pattern; site selector via shadcn `Select`; ad-size + format selects; details route shows zone stats + embed code.
4. **Ad Code Generator** — zone `Select`, toggles for async/responsive, textarea preview of JS snippet, copy-to-clipboard w/ toast, live placeholder box mimicking the chosen ad size.
5. **Statistics** — rewrite existing `_app.stats.tsx`: date range presets + custom range (shadcn `Calendar` in popover), metric cards, tabs for Site/Zone/Country/Device/Browser/OS, three line charts, sortable paginated table with CSV export (client-side blob download).
6. **Payouts** — three balance cards, payout history table, request payout modal, payment methods list w/ add/edit modal (bank + mobile money), net-30 info banner, tax info upload dropzone (client stub).
7. **Settings** — tabs (Profile, Security, Payments, Notifications, API keys, 2FA); toggles use shadcn `Switch`; API keys list with generate + revoke buttons.
8. **Reports** — performance/geo/device/traffic tabs, date range + metric multi-select, scheduled reports table with add-schedule modal.

## Auth flow

- `/auth/login` — email + password form, POST assumed `/api/auth/login` returning `{ token, user }`; store in localStorage; redirect to `/`.
- Logout button in sidebar footer (replacing the top-up link once implemented) clears token + navigates to login.
- All `_app/*` routes gated; loaders that hit the API surface errors via existing `errorComponent` patterns.

## Technical notes

- Keep the existing store `src/lib/store.ts` file but repurpose it (or add `publisherStore`) to cache balance + user profile for the sidebar widget.
- All new fetches go through TanStack Query with keys like `["sites"]`, `["zones", siteId?]`, `["stats", range]`, `["balance"]`, `["payouts"]`.
- CORS: the API is on plain HTTP at an IP. If the browser blocks mixed content or CORS on the deployed HTTPS preview, we may need a small server route proxy under `src/routes/api/proxy.$.ts` — I'll add this if the direct fetch fails during verification. (Optional; only if needed.)
- No design token changes, no font changes, no new color additions. Chart colors reuse the two greens already in `_app.index.tsx`.
- Head metadata: update `__root.tsx` title/description to "PropelAds Publisher — Monetize your traffic" and matching og tags.

## Out of scope

- Real-time websocket updates (dashboard uses polling `refetchInterval: 30_000`).
- Backend implementation — we consume the given API as-is.
- Dark mode toggle — the current design is a single light theme; adding a toggle would require new tokens. I'll note this and skip unless you confirm you want it now.
- Lovable Cloud — not enabling since you specified an external API + JWT.

## Deliverable order

1. Auth (login page, api client, token store, route gate)
2. Sidebar/branding update
3. Dashboard
4. Sites + Zones (with details subroutes)
5. Ad Code Generator
6. Statistics rewrite
7. Payouts
8. Reports
9. Settings
10. Cleanup: remove advertiser routes, update root head metadata, verify build

Ready to proceed on approval.
