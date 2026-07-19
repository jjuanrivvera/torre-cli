<div align="center">

# torre

[![CI](https://github.com/jjuanrivvera/torre-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/jjuanrivvera/torre-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/jjuanrivvera/torre-cli)](https://github.com/jjuanrivvera/torre-cli/releases/latest)
[![Coverage](https://img.shields.io/badge/coverage-%E2%89%A580%25-brightgreen)](https://github.com/jjuanrivvera/torre-cli/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/jjuanrivvera/torre-cli.svg)](https://pkg.go.dev/github.com/jjuanrivvera/torre-cli)
[![Go version](https://img.shields.io/github/go-mod/go-version/jjuanrivvera/torre-cli)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/jjuanrivvera/torre-cli)
[![Built with cliwright](https://img.shields.io/badge/built_with-cliwright-1f6feb)](https://cliwright.jjuanrivvera.com)

**Torre.ai jobs from your terminal — search opportunities, fetch public genomes, agent-friendly output (JSON/YAML/CSV/MCP).**

[Documentation](https://jjuanrivvera.github.io/torre-cli/) · [Command reference](https://jjuanrivvera.github.io/torre-cli/commands/torre/)

</div>

A fast, scriptable, **agent-friendly** command-line client for the [Torre.ai](https://torre.ai)
public API. Search job opportunities, fetch an opportunity's full detail, search people, and pull
a person's public genome/bio — all with machine-first output (JSON/YAML/CSV, `-o id`, `--jq`) so an
AI assistant or a shell pipeline can consume it directly.

Torre's public endpoints need **no credentials**, so `torre` works out of the box.

## Install

**Install script (macOS/Linux)** — checksum-verified, no dependencies:

```sh
curl -fsSL https://raw.githubusercontent.com/jjuanrivvera/torre-cli/main/install.sh | sh
```

**Homebrew:**

```sh
brew install jjuanrivvera/torre-cli/torre-cli
```

**Go:**

```sh
go install github.com/jjuanrivvera/torre-cli/cmd/torre@latest
```

**Windows (Scoop):**

```powershell
scoop bucket add torre https://github.com/jjuanrivvera/scoop-torre-cli
scoop install torre
```

## Quickstart

```sh
# Search remote Go roles, machine-readable
torre jobs search --skill golang --remote -o json

# LATAM + remote, capped and piped
torre jobs search --skill "product design" --location Colombia --limit 50 -o id

# One opportunity's full detail
torre jobs get KWN4QjAd

# A person's public genome, sliced with jq
torre genome torrenegra --jq '.person.name'
torre genome torrenegra --jq '[.strengths[].name]' -o json

# Search people, then pull a match's profile
torre people search --skill "data science" --remote -o table
```

## Filters (jobs search)

| Flag | Meaning |
|---|---|
| `--skill` / `--query` | skill or role text |
| `--experience` | experience level (default `potential-to-develop`; Torre's seniority proxy — e.g. `1-plus-years`, `3-plus-years`) |
| `--remote` | only remote opportunities |
| `--location` | location/country (e.g. `Colombia`) |
| `--organization` | organization name |
| `--compensation` `--currency` `--periodicity` | minimum compensation (currency default `USD$`, periodicity default `monthly`) |
| `--size` `--limit` `--all` | pagination |

## Output & scripting

Global `-o table|json|yaml|csv|id`, `--columns`, `--jq <gojq>`, `--dry-run` (prints the
equivalent `curl`), `-v/--verbose`, `--no-color`/`NO_COLOR`. Notes go to stderr so stdout stays
pipe-clean. CSV cells are sanitized against spreadsheet formula injection.

## For AI agents

- **MCP server:** `torre mcp start` exposes the read commands as annotated MCP tools; setup/secret
  commands and the raw `api` escape hatch are excluded.
- **Agent guard:** `torre agent guard --host claude-code|codex|opencode` emits host safety config
  from the live command tree (torre is read-only, so the guard mainly gates the raw `api` escape
  hatch to GET/HEAD/OPTIONS and blocks `alias set`).

## Auth (optional)

Every wrapped endpoint is public. If you have a bearer token for an authenticated endpoint:

```sh
torre auth login          # hidden prompt; stored in the OS keyring
export TORRE_TOKEN=...     # or per-invocation via env
```

## Honest comparison

There is no official Torre CLI. `torre` wraps the same public endpoints the Torre web app uses;
it adds machine output, an MCP server, agent guardrails, and offline-friendly scripting. It does
**not** cover authenticated/private Torre features (applications, messaging) — use the web app for
those.

## License

MIT — see [LICENSE](LICENSE).
