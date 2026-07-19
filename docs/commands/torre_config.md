## torre config

Inspect and edit torre configuration

### Synopsis

The config file holds only non-secret settings (profiles, host overrides, aliases).
Any bearer token lives in the OS keyring — never here.

### Options

```
  -h, --help   help for config
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
* [torre config list-profiles](torre_config_list-profiles.md)	 - List configured profiles
* [torre config path](torre_config_path.md)	 - Print the config file path
* [torre config set](torre_config_set.md)	 - Set a per-profile option (search_base_url, api_base_url)
* [torre config use](torre_config_use.md)	 - Set the default profile for future invocations
* [torre config view](torre_config_view.md)	 - Show the resolved configuration

