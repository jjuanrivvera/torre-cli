## torre update

Update torre to the latest GitHub release

### Synopsis

Download the latest torre release, verify it against checksums.txt, and replace
the running binary in place. Use 'torre update check' to see what's available without
installing.

```
torre update [flags]
```

### Examples

```
  torre update
  torre update check
```

### Options

```
  -h, --help   help for update
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
* [torre update check](torre_update_check.md)	 - Check for a newer release without installing it

