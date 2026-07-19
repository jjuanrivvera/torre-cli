## torre jobs search

Search job opportunities

### Synopsis

Search Torre.ai opportunities with skill/role, remote, location, organization, and
compensation filters. Results paginate with --size/--limit/--all. Machine output
(-o json/-o id/--jq) is the primary interface for an assistant; -o table is the human view.

A skill search needs an experience level (Torre rejects a bare skill); --experience defaults
to "potential-to-develop" and accepts Torre's levels such as "1-plus-years",
"2-plus-years", "3-plus-years", "5-plus-years".

```
torre jobs search [flags]
```

### Examples

```
  torre jobs search --skill golang --remote
  torre jobs search --skill "product design" --location Colombia --limit 50 -o json
  torre jobs search --skill go --compensation 3000 --currency 'USD$' --periodicity monthly
  torre jobs search --skill go --remote -o id | head
```

### Options

```
      --compensation float    minimum compensation amount
      --currency string       compensation currency (default "USD$")
      --experience string     required experience level (default potential-to-develop)
  -h, --help                  help for search
      --location string       location/country to match (e.g. Colombia)
      --organization string   organization name to match
      --periodicity string    compensation periodicity: hourly|monthly|yearly (default monthly)
      --remote                only remote opportunities
      --skill string          skill or role text to match
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

* [torre jobs](torre_jobs.md)	 - Search and inspect Torre job opportunities

