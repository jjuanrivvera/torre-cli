## torre config set

Set a per-profile option (search_base_url, api_base_url)

### Synopsis

Set a non-secret host override on the ACTIVE profile (--profile selects it).
Keys: search_base_url (default https://search.torre.co), api_base_url
(default https://torre.ai/api).

```
torre config set <key> <value> [flags]
```

### Examples

```
  torre config set api_base_url https://torre.ai/api --profile default
```

### Options

```
  -h, --help   help for set
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

* [torre config](torre_config.md)	 - Inspect and edit torre configuration

