## torre people search

Search people by skill/role

### Synopsis

Search Torre.ai people. A skill search needs an experience level (Torre rejects a
bare skill); --experience defaults to "potential-to-develop". Pair a match with
`torre genome <username>` to pull a candidate's full profile.

```
torre people search [flags]
```

### Examples

```
  torre people search --skill "data science" --remote
  torre people search --skill golang --location Colombia --limit 25 -o json
```

### Options

```
      --experience string   required experience level (default potential-to-develop)
  -h, --help                help for search
      --location string     location/country to match
      --remote              only people open to remote work
      --skill string        skill or role text to match
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

* [torre people](torre_people.md)	 - Search Torre people

