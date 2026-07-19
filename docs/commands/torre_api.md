## torre api

Send a raw Torre request (escape hatch)

### Synopsis

Call any Torre endpoint directly. --host selects which base the PATH is relative to:
"api" (default, https://torre.ai/api — genome, opportunity detail) or "search"
(https://search.torre.co — the _search endpoints).

This is the documented escape hatch for anything torre does not wrap as a first-class
command. It honors --dry-run, -o/--output, and --jq like every other command. Non-GET
methods are never auto-retried.

```
torre api <METHOD> <PATH> [--host search|api] [-d body] [-q key=value ...] [flags]
```

### Examples

```
  torre api GET genome/bios/torrenegra
  torre api GET suite/opportunities/KWN4QjAd
  torre api POST opportunities/_search/ --host search -q size=5 -d '{"and":[{"remote":{"term":true}}]}'
```

### Options

```
  -d, --data string         JSON body: inline, @file, or - for stdin
  -h, --help                help for api
      --host string         which Torre host: api|search (default "api")
  -q, --query stringArray   query parameter key=value (repeatable)
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

* [torre](torre.md)	 - A fast, scriptable CLI for Torre.ai jobs and profiles

