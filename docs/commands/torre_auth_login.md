## torre auth login

Store a Torre bearer token in the OS keyring

### Synopsis

Capture a bearer token for the active profile and store it in the OS keyring
(encrypted-file fallback on headless hosts, keyed by $TORRE_KEYRING_PASSWORD). The token is
read from a hidden prompt so it never echoes to the terminal; use --token only in trusted,
non-interactive contexts (it can leak into shell history).

```
torre auth login [flags]
```

### Examples

```
  torre auth login
  torre auth login --profile work
```

### Options

```
  -h, --help           help for login
      --token string   token value (prefer the hidden prompt; --token can leak into history)
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

* [torre auth](torre_auth.md)	 - Manage an optional Torre bearer token

