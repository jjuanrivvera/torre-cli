# AGENTS.md — working in the torre-cli repo

`torre` is a read-only, agent-friendly command-line client for the **Torre.ai public API**:
search job opportunities, fetch an opportunity's detail, search people, and pull a person's
public genome/bio. Built to the cliwright standard (Go + Cobra + GoReleaser). This file
orients an AI agent (or human) contributing.

## The one rule that matters
**`make verify` is the gate.** A change is done only when `make verify` exits `0`. It runs
`make check` (fmt, vet, golangci-lint, tests) + `spec-check` (built surface == `api-manifest.json`)
+ `spec-completeness` (manifest wraps ≥90% of the 4-endpoint enumerated public API — currently
100%) + `cover-check` (≥80% coverage) + `dod-check.sh`. Run the full `make verify` for any change
that touches the command surface or a documented behavior — not just `make check`.

## Architecture (where things live)
- `internal/api/` — the Torre client core. **Two hosts, one client**: the search cluster
  (`search.torre.co`, POST `_search`) and the app API (`torre.ai/api`, GET detail + genome);
  `NewClientWithBaseURL` points both at one URL for tests. Idempotent-only retry honoring
  `Retry-After` with full-jitter backoff, dry-run curl with token redaction, `APIError` with
  actionable hints, flexible JSON types (`ID`/`Int`/`Bool`/`StringOrSlice`), and typed service
  methods (`SearchOpportunities`, `SearchOpportunitiesAll`, `GetOpportunity`, `SearchPeople`,
  `SearchPeopleAll`, `Genome`). **Pattern B (service-layer)** — read-only, non-CRUD endpoints
  (DECISIONS.md #10).
- `internal/auth/` — OPTIONAL bearer token in the OS keyring (service `torre-cli`, key
  `profile-<name>`), AES-256-GCM encrypted-file fallback (`TORRE_KEYRING_PASSWORD`). Torre's
  public endpoints need no token; a token is only for endpoints that require one.
- `commands/` — the cobra tree. `init()` appends builders to `registrars`/`metaRegistrars`;
  `NewRootCmd()` drains the queue onto a fresh root. MCP annotations are stamped via
  `annotate(cmd, kind)` as commands are built (everything is `kindRead`).
- `internal/{config,output,version,update}` — profiles + manual precedence (no Viper), the
  table/json/yaml/csv/id renderer (CSV formula-injection guard, terminal-escape sanitizer,
  NO_COLOR), build metadata, the checksum-verified self-updater.
- `cmd/torre/main.go` — `signal.NotifyContext` (Ctrl-C cancels pagination + retry backoff) +
  alias expansion before cobra parses.

## Torre specifics you must not re-derive (all verified live 2026-07-18; see DECISIONS.md)
- Opportunity search: `POST https://search.torre.co/opportunities/_search/` (trailing slash),
  `?size&offset&aggregate=false`, body `{"and":[<clause>...]}`. Response `{total,size,results}`.
- Opportunity detail: `GET https://torre.ai/api/suite/opportunities/{id}`.
- Genome: `GET https://torre.ai/api/genome/bios/{username}`.
- People search: `POST https://search.torre.co/people/_search`.
- A skill search REQUIRES an experience level; `--experience` defaults to `potential-to-develop`
  (Torre models seniority via this level). Compensation currency literal is `USD$`.
- Public/no-auth is the target path — never block it on a missing token.

## House rules
- Comments explain **WHY**, not WHAT.
- Thread `cmd.Context()` everywhere; never `context.Background()` (it breaks Ctrl-C).
- Any token lives in the OS keyring — never in config, code, or commit messages.
- Pin every ambiguous API assumption in `DECISIONS.md`; read it back, never re-decide.
- Surface changes require updating `api-manifest.json` AND regenerating docs
  (`make docs-gen`) in the same commit.
- MCP exclusions are by EXACT path (`commands/mcp.go`), never substring.
- New commands ship with tests in the same commit — coverage is a ratchet (≥80%).
