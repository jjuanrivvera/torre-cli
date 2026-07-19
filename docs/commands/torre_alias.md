## torre alias

Manage user-defined command aliases

### Synopsis

Define shorthand commands. Aliases are expanded before parsing and can never shadow a built-in.

### Options

```
  -h, --help   help for alias
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
* [torre alias list](torre_alias_list.md)	 - List aliases
* [torre alias remove](torre_alias_remove.md)	 - Remove an alias
* [torre alias set](torre_alias_set.md)	 - Create or update an alias

