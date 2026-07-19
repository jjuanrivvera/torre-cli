# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.2] - 2026-07-19

### Fixed
- **`jobs search` pagination duplication** — `--limit N`/`--all` returned N rows but only a
  single page's worth of UNIQUE ids (exactly 5× duplication at `--limit 100`). Root cause: the
  opportunities `_search` endpoint **silently ignores the `offset` query parameter** (verified
  live 2026-07-19: `offset=5` still returns `offset:0` and identical ids), so the offset loop
  re-fetched page 1 every iteration. Fixed on two layers: (1) the paginator now requests the
  largest single page the server allows (`size` capped at 99 — `size>=100` is rejected HTTP
  400) instead of a dead offset walk, and stops as soon as a page yields no new id; (2) results
  are **de-duplicated by `.id`** while accumulating, so the CLI never emits a duplicate
  opportunity even if the server returns overlapping windows.

### Added
- `jobs search --location-type <value>` (repeatable/CSV) and `--remote-anywhere` — hard
  client-side filters over each opportunity's `.place.locationType` (case-insensitive). For a
  remote LATAM/Colombian contractor, `--remote-anywhere` (shorthand for
  `--location-type remote_anywhere`) keeps only roles open to any country — a stronger quality
  filter than the soft `--location` ranking hint. Opportunities with a missing/empty
  locationType are dropped when the filter is active.
- `jobs search --comp-disclosed-only` — hard client-side filter keeping only opportunities that
  disclose a compensation figure (`.compensation.data.minAmount > 0` or `minHourlyUSD > 0`).
  Distinct from the `--compensation` ranking hint. Both new filters compose (AND) with
  `--since`, and — like `--since` — widen the scan when active without an explicit `--limit`.

### Changed
- Docs: clarify that `jobs search --location` and `--compensation` (with `--currency`/
  `--periodicity`) are **server-side ranking hints, not hard filters** — Torre's `_search`
  endpoint treats them as relevance boosts and does not restrict results to that location or
  pay (verified 2026-07-19: `--skill go`, `--skill go --location Colombia` and
  `--skill go --compensation 4000` return identical counts). Flag help text, the `jobs search`
  long description, the README cheatsheet and the skill are reworded accordingly. `--since`
  remains the one hard (client-side) filter. No behavior change.

## [0.1.1] - 2026-07-19

### Added
- `jobs search --since` (alias `--posted-after`) — a posting-date filter. Accepts an absolute
  date `YYYY-MM-DD` or a relative shorthand `Nd`/`Nw` (last N days/weeks) and keeps only
  opportunities whose `.created` timestamp is on/after the threshold. Torre orders search
  results by relevance rather than date, so the filter is applied client-side over the fetched
  page(s); when a date is pinned without an explicit `--limit`/`--all` the scan is widened so
  sparse recent items still surface. Pair with `--all` or a larger `--limit`.

## [0.1.0] - 2026-07-18

### Added
- Initial release of `torre`, a read-only agent-friendly CLI for the Torre.ai public API.
- `jobs search` — search opportunities with skill/role, remote, location, organization, and
  compensation filters; `--size`/`--limit`/`--all` pagination.
- `jobs get <id>` — fetch one opportunity's full detail.
- `genome <username>` — fetch a person's public genome/bio.
- `people search` — search the Torre people index.
- Output formats: table, json, yaml, csv, `-o id`, `--jq`, `--columns`, `--dry-run` curl.
- Optional bearer-token auth stored in the OS keyring (AES-GCM encrypted-file fallback).
- MCP server (`mcp`), agent guard (`agent guard`), and the standard meta command set
  (`auth`, `config`, `init`, `doctor`, `completion`, `alias`, `api`, `version`, `update`).
