## torre jobs search

Search job opportunities

### Synopsis

Search Torre.ai opportunities with skill/role, remote, location, organization, and
compensation filters. Results paginate with --size/--limit/--all. Machine output
(-o json/-o id/--jq) is the primary interface for an assistant; -o table is the human view.

Not every flag narrows the result set. --skill (and --experience) narrows the search, and
--since (alias --posted-after), --location-type/--remote-anywhere and --comp-disclosed-only
are hard client-side filters. But --location and --compensation (with --currency/--periodicity)
are RANKING HINTS Torre applies server-side: they nudge relevance/ordering, they do NOT
restrict results to that location or pay. A remote role, for example, carries no location and
is not dropped by --location.

For a remote contractor, --remote-anywhere (shorthand for --location-type remote_anywhere)
keeps only roles open to any country — a stronger quality filter than the soft --location.
--comp-disclosed-only drops opportunities that hide their pay. Both compose (AND) with --since
over the correctly-paginated, de-duplicated result set.

A skill search needs an experience level (Torre rejects a bare skill); --experience defaults
to "potential-to-develop" and accepts Torre's levels such as "1-plus-years",
"2-plus-years", "3-plus-years", "5-plus-years".

Results come back ordered by RELEVANCE, not date, and span years. --since (alias
--posted-after) drops anything older than a threshold — an absolute YYYY-MM-DD or a relative
Nd/Nw (last N days/weeks). Because recent items are sparse in a small relevance-ordered page,
pair --since with --all or a larger --limit; when neither is set --since widens the scan.

```
torre jobs search [flags]
```

### Examples

```
  torre jobs search --skill golang --remote
  torre jobs search --skill "product design" --location Colombia --limit 50 -o json
  torre jobs search --skill go --since 7d --remote -o json
  torre jobs search --skill go --remote-anywhere --limit 100 -o id
  torre jobs search --skill go --location-type remote_anywhere,remote_timezones -o json
  torre jobs search --skill go --comp-disclosed-only --since 14d -o json
  torre jobs search --skill go --posted-after 2026-07-12 --all -o id
  torre jobs search --skill go --compensation 3000 --currency 'USD$' --periodicity monthly
```

### Options

```
      --comp-disclosed-only     hard client-side filter: keep only opportunities that disclose a compensation figure (minAmount>0 or minHourlyUSD>0); distinct from the --compensation ranking hint
      --compensation float      compensation ranking hint applied server-side (nudges relevance/ordering; does NOT restrict results)
      --currency string         currency for the --compensation ranking hint (default "USD$")
      --experience string       required experience level (default potential-to-develop)
  -h, --help                    help for search
      --location string         location/country ranking hint applied server-side (nudges relevance/ordering; does NOT restrict results — unlike --since)
      --location-type strings   hard client-side filter: keep only opportunities whose .place.locationType matches (case-insensitive, repeatable/CSV; e.g. remote_anywhere,remote_timezones)
      --organization string     organization name to match
      --periodicity string      periodicity for the --compensation ranking hint: hourly|monthly|yearly (default monthly)
      --remote                  only remote opportunities
      --remote-anywhere         shorthand for --location-type remote_anywhere (roles open to any country — the key filter for a remote LATAM contractor)
      --since string            keep only opportunities created on/after this date: absolute YYYY-MM-DD or relative Nd/Nw (e.g. 7d, 2w)
      --skill string            skill or role text to match
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

