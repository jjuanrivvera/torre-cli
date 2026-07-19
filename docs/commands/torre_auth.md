## torre auth

Manage an optional Torre bearer token

### Synopsis

Torre's public endpoints (job search, opportunity detail, genome, people search)
need no credentials, so torre works with no auth at all. A bearer token is only useful for
endpoints that require one; store it here and it lands in your OS keyring, scoped to the
active profile. You can also pass a token per-invocation via the TORRE_TOKEN env var.

### Options

```
  -h, --help   help for auth
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
* [torre auth login](torre_auth_login.md)	 - Store a Torre bearer token in the OS keyring
* [torre auth logout](torre_auth_logout.md)	 - Remove the stored token for the active profile
* [torre auth status](torre_auth_status.md)	 - Show the active profile and whether a token is stored

