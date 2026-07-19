## torre genome

Fetch a person's public Torre genome/bio

### Synopsis

Fetch a person's public genome (their Torre bio: profile, strengths, experiences,
education, interests, and more) by username — the handle in their profile URL
(torre.ai/<username>). The full object is large; use --jq or -o json to slice it, which is
ideal for an assistant computing a candidate/role match.

```
torre genome <username> [flags]
```

### Examples

```
  torre genome torrenegra
  torre genome torrenegra --jq '.person.name'
  torre genome torrenegra --jq '[.strengths[].name]' -o json
  torre genome torrenegra -o yaml
```

### Options

```
  -h, --help   help for genome
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

