## torre agent guard

Generate agent-safety config that blocks destructive torre operations

### Synopsis

Classify every API command (read / write / irreversible) from the live command tree
and emit host safety config. torre is a READ-ONLY client, so today every command is a
read and the guard's main job is to gate the raw "torre api <METHOD> <PATH>" escape hatch
(only GET/HEAD/OPTIONS pass) and "torre alias set". The guard derives from the LIVE tree,
so if a future write/destructive command is ever added it is hard-blocked automatically,
including its cobra alias paths.

For claude-code the output also includes a PreToolUse hook script
(.claude/hooks/torre-guard.sh): it strips quote/backslash obfuscation, matches blocked
subcommand paths at the command position even for path-invoked binaries (./bin/torre,
/usr/local/bin/torre), and gates the raw "torre api <METHOD> <PATH>" escape hatch at
the METHOD position — only GET/HEAD/OPTIONS pass; POST/PUT/PATCH/DELETE are denied
case-insensitively, while a GET whose path merely contains "delete" stays allowed.
"torre alias set" is denied so an agent cannot mint a new shorthand for a blocked
command.

MCP-only operation is the hard guarantee; the Bash rails are best-effort — the hook
defeats quoting tricks and path prefixes, but not variable indirection
(a=DELETE; torre api $a x) or shell aliases. Conservative false positives are
accepted: a line that merely QUOTES a blocked command is denied.

```
torre agent guard --host <claude-code|codex|opencode> [flags]
```

### Examples

```
  torre agent guard --host claude-code
  torre agent guard --host claude-code --write          # write the files into .claude/
  torre agent guard --host codex --out ~/.codex/config.toml
  torre agent guard --host opencode --all-writes
```

### Options

```
      --all-writes    also hard-block ordinary writes, not just irreversible ops
  -h, --help          help for guard
      --host string   target agent host: claude-code|codex|opencode (required)
      --out string    write to this file instead of stdout
      --write         claude-code only: write hook + settings fragment under .claude/ (never overwrites)
```

### Options inherited from parent commands

```
      --all                      page through all results (search commands)
      --base-url string          override the Torre app-API host (default https://torre.ai/api)
      --columns strings          comma-separated columns to show
      --dry-run                  print the equivalent curl and make no request
      --jq string                gojq expression applied to the response before rendering
      --limit int                max items to return across pages (search commands)
      --no-color                 disable colored output
  -o, --output string            output format: table|json|yaml|csv|id
      --profile string           named profile to use
      --quiet                    suppress non-essential chatter
      --search-base-url string   override the Torre search host (default https://search.torre.co)
      --show-token               reveal the bearer token in dry-run output
      --size int                 results per page (search commands) (default 20)
  -v, --verbose                  verbose request logging (stderr)
```

### SEE ALSO

* [torre agent](torre_agent.md)	 - AI-agent integration helpers

