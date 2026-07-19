## torre

A fast, scriptable CLI for Torre.ai jobs and profiles

### Synopsis

torre is a read-only, agent-friendly client for the Torre.ai public API: search job
opportunities, fetch an opportunity's detail, search people, and pull a person's public
genome/bio — all with machine-first output (JSON/YAML/CSV, -o id, --jq) so an AI assistant
or a shell pipeline can consume it directly.

Torre's public endpoints need no credentials, so it works out of the box. A bearer token is
optional (torre auth login) for any endpoint that requires one.

Examples:
  torre jobs search --skill "golang" --remote -o json
  torre jobs search --skill "product design" --location Colombia --limit 50
  torre jobs get KWN4QjAd
  torre genome torrenegra --jq '.person.name'
  torre people search --skill "data science" --remote -o table

### Options

```
      --all                      page through all results (search commands)
      --base-url string          override the Torre app-API host (default https://torre.ai/api)
      --columns strings          comma-separated columns to show
      --dry-run                  print the equivalent curl and make no request
  -h, --help                     help for torre
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

* [torre agent](torre_agent.md)	 - AI-agent integration helpers
* [torre alias](torre_alias.md)	 - Manage user-defined command aliases
* [torre api](torre_api.md)	 - Send a raw Torre request (escape hatch)
* [torre auth](torre_auth.md)	 - Manage an optional Torre bearer token
* [torre completion](torre_completion.md)	 - Generate shell completion scripts
* [torre config](torre_config.md)	 - Inspect and edit torre configuration
* [torre doctor](torre_doctor.md)	 - Diagnose configuration, keyring, and Torre connectivity
* [torre genome](torre_genome.md)	 - Fetch a person's public Torre genome/bio
* [torre init](torre_init.md)	 - First-run setup wizard
* [torre jobs](torre_jobs.md)	 - Search and inspect Torre job opportunities
* [torre mcp](torre_mcp.md)	 - MCP server management
* [torre people](torre_people.md)	 - Search Torre people
* [torre update](torre_update.md)	 - Update torre to the latest GitHub release
* [torre version](torre_version.md)	 - Print version, commit, and build date

