# DECISIONS — pinned Torre.ai API assumptions

One line each: question → decision → why. Read back every iteration; never silently re-decide.
All endpoints below were verified by live probe on 2026-07-18 (Torre publishes no official
OpenAPI/llms.txt/Postman collection, so the surface was enumerated by request).

## Endpoints (verified live 2026-07-18)

1. **Opportunity search host/path** → `POST https://search.torre.co/opportunities/_search/`
   (trailing slash required) with query params `size`, `offset`, `aggregate=false`, and a JSON
   boolean-query body. Why: returned HTTP 200 with `{"total":N,"size":N,"results":[...]}`;
   the newer `arda.torre.co/opportunities/_search` host 404s, so `search.torre.co` is the live one.
2. **Opportunity detail** → `GET https://torre.ai/api/suite/opportunities/{id}`. Why: returned
   HTTP 200 with the full opportunity object; `torre.ai/api/opportunities/{id}` and the
   `search.torre.co` detail variants all 404.
3. **Genome/bio** → `GET https://torre.ai/api/genome/bios/{username}`. Why: returned HTTP 200
   (`bio.torre.co/api/bios/{username}` is an equivalent alias; `torre.ai/api/bios/...` 404s).
4. **People search** → `POST https://search.torre.co/people/_search` with `size`/`offset`/
   `aggregate`. Why: returned HTTP 200 with the same `{total,size,results}` envelope.

## Search query DSL

5. **Filter body shape** → opportunities use `{"and":[<clause>...]}`; people use a flat object.
   Clauses verified live: `{"skill/role":{"text":T,"experience":E}}`, `{"remote":{"term":true}}`,
   `{"location":{"term":L}}`, `{"compensation":{"value":V,"currency":C,"periodicity":P}}`,
   `{"organization":{"term":O}}`. An empty body returns the full firehose.
6. **A skill search REQUIRES an experience level** → default `--experience` to
   `potential-to-develop` when a skill is given. Why: people `_search` returns HTTP 500
   ("Either experience or proficiency should be provided in the skill/role query") for a bare
   `text`; applying a default keeps the common case working. Torre models "seniority" through
   this experience level (`1-plus-years`, `2-plus-years`, `3-plus-years`, `5-plus-years`,
   `potential-to-develop`), so `--experience` is the seniority knob rather than a separate field.
7. **Compensation currency literal** → default `USD$` (Torre's literal token, dollar suffix
   included), periodicity default `monthly`. Why: the verified body used `"currency":"USD$"`.

## Auth / architecture

8. **Auth is OPTIONAL** → all wrapped endpoints are public and unauthenticated; a bearer token
   is supported (env `TORRE_TOKEN` or keyring via `auth login`) but never required. Why: every
   endpoint returned 200 with no `Authorization` header during recon. The public path must never
   be blocked on missing credentials.
9. **Two hosts, one client** → the client holds both a search base (`search.torre.co`) and an
   app-API base (`torre.ai/api`) and routes each method to the right one; `NewClientWithBaseURL`
   points both at one URL for tests. Why: Torre's public surface genuinely spans two hosts — a
   single base can't address both.
10. **Pattern B (service-layer), not generic-core** → the endpoints are read-only and non-CRUD
    (POST-with-query-DSL search, GET-by-id, GET-by-username), which is the documented §11 trigger
    for Pattern B. Typed service methods (`SearchOpportunities`, `GetOpportunity`, `SearchPeople`,
    `Genome`) render raw JSON through the shared formatter.
11. **Default User-Agent** → a browser-ish UA is sent on every request. Why: Torre's edge is
    friendlier to a non-bare-Go UA; overridable via `TORRE_USER_AGENT`.
12. **POST searches are not auto-retried** → per the idempotent-only retry rule (§1); GET detail
    and genome are retried. Why: safe default even though searches are semantically read-only.

## Conditional patterns (§3d) — decisions

- Event-store / offline-cache: **N/A** — Torre is a stateless read API with its own durable
  search; no ephemeral stream and no need for a local system-of-record.
- Spec-contract test / smoke.yml / spec-sync.yml: **N/A** — no machine spec exists at a stable
  URL to diff against; drift is caught by the recon-derived DECISIONS + manifest.
- Multi-group credentials: **N/A** — no auth tiers (public API).
- Adopt a typed library: **N/A** — no mature Go client library for Torre.
- Terminal-escape sanitization: **applied** (shared renderer) — job titles/names are free text.
- CSV: **kept** — search results are tabular enough to be useful; genome is not, so it renders
  best as json/yaml.

## Posting-date filter (`--since`)

13. **Date filtering is CLIENT-SIDE, not server-side** → `jobs search --since`/`--posted-after`
    filters the fetched result set on each opportunity's `.created` field; the search body is
    unchanged. Why: verified live 2026-07-19 that the opportunities `_search` endpoint has no
    documented date clause and **silently ignores** one — probing `{"and":[{"created":{"from":
    "2026-07-01"}}]}`, `{"created":{"gte":...}}`, `{"created":{"term":{"gte":...}}}` and
    `{"date":{"from":...}}` all returned the same full `total` (266040) as an empty body, i.e.
    no filtering happened. Results are ordered by RELEVANCE, not date, and span 2021→today, so a
    client-side pass over `.created` (RFC 3339, e.g. `2025-04-30T16:42:17.000Z`) is the only
    reliable date filter. If Torre later ships a real created clause, prefer server-side and
    revise this note.
14. **`--since` widens the scan when no explicit `--limit`/`--all`** → because relevance ordering
    buries recent items in a small page, pinning `--since` without an explicit `--limit` and
    without `--all` bumps the fetch to `sinceDefaultScan` (100). Why: a default single page of 20
    relevance-ranked results often contains zero recent items; scanning more finds them without
    forcing an unbounded `--all`. Documented on the flag and in the README/SKILL cheatsheets.
15. **`--since` scopes to `jobs search` only** → people results carry no comparable created/date
    field (verified 2026-07-19: a people `_search` result exposes `ggId`, `name`, `username`,
    `professionalHeadline`, `verified`, `weight`, … but no timestamp), so a date filter is
    inapplicable there.

## Completeness

- `api_method_total = 4` is the full enumerated public unauthenticated surface (see
  `enumerated_endpoints`). The manifest covers all 4 (jobs search/get, people search, genome),
  so `make spec-completeness` reports 100%. No coverage-waiver needed.
