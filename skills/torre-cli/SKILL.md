---
name: torre-cli
description: Use this when you need to discover jobs or analyze candidate/role fit on Torre.ai — search opportunities by skill/role, remote, location, organization, and compensation, fetch a single opportunity, or pull a person's public genome (bio, strengths, experiences) to compute a match. Read-only; no account or token needed.
version: 0.1.0
homepage: https://github.com/jjuanrivvera/torre-cli
license: MIT
allowed-tools: Bash(torre:*)
metadata: {"openclaw":{"category":"jobs","emoji":"💼","requires":{"bins":["torre"],"env":[]},"install":[{"kind":"brew","formula":"jjuanrivvera/torre-cli/torre-cli","bins":["torre"]},{"kind":"go","package":"github.com/jjuanrivvera/torre-cli/cmd/torre@latest","bins":["torre"]}]}}
---

# torre — Torre.ai jobs CLI

`torre` is a read-only client for the [Torre.ai](https://torre.ai) public API. Prefer it over
raw `curl`: it targets the live hosts (`search.torre.co` for search, `torre.ai/api` for
detail/genome), builds the opportunity-search request body correctly, and gives you clean
JSON/`-o id` output plus a built-in `--jq`.

## Prerequisites
- Install: `brew install jjuanrivvera/torre-cli/torre-cli` or
  `go install github.com/jjuanrivvera/torre-cli/cmd/torre@latest`.
- No account or token needed for job search — every wrapped endpoint is public. (An optional
  `TORRE_TOKEN` exists for higher limits but never blocks the public path.) `torre doctor`
  verifies connectivity.

## Golden rules
1. **A skill search needs an experience level.** Torre rejects a bare skill, so `--experience`
   defaults to `potential-to-develop`; other values: `1-plus-years`, `2-plus-years`,
   `3-plus-years`, `5-plus-years`.
2. **Emit machine output** for downstream steps: `-o json`, `-o id`, or slice with `--jq`.
3. **Filters compose:** `--skill`, `--remote`, `--location`, `--organization`,
   `--compensation`/`--currency`/`--periodicity`; page with `--limit`/`--size`/`--all`.
4. **`genome` is large** — always `--jq` or `-o json` a slice; it's ideal for computing a
   candidate/role match against a profile.

## Workflow (discover → inspect → match)

```sh
# 1. Discover — recent remote Go roles, as JSON
torre jobs search --skill golang --remote --limit 20 -o json

# 2. Narrow by location, organization, or pay
torre jobs search --skill "backend" --location Colombia -o json
torre jobs search --skill go --compensation 3000 --currency 'USD$' --periodicity monthly -o json

# 3. Inspect one opportunity
torre jobs get <opportunity-id> -o json

# 4. Pull a person's genome to compute fit
torre genome <username> --jq '{name:.person.name, skills:[.strengths[].name]}'
```

## Cheatsheet

| Task | Command |
|---|---|
| Remote roles by skill | `torre jobs search --skill <skill> --remote --limit 20` |
| By location | `torre jobs search --skill <skill> --location Colombia` |
| By organization | `torre jobs search --skill <skill> --organization <org>` |
| Min compensation | `torre jobs search --skill <skill> --compensation 3000 --currency 'USD$'` |
| One opportunity | `torre jobs get <id> -o json` |
| Just ids | `torre jobs search --skill <skill> -o id` |
| Person's genome | `torre genome <username> -o json` |
| People search | `torre people search --skill <skill> --location Colombia` |
| See the request | `torre jobs search --skill go --dry-run` |

## Troubleshooting
- **Empty results for a skill:** you likely omitted `--experience` context — the default is
  applied automatically, but a very narrow `--location`/`--compensation` can zero out results.
- **Genome too big:** slice it with `--jq` rather than dumping the whole object.
- **Connectivity:** `torre doctor` checks config, both hosts, and a live request.
