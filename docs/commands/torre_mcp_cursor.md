## torre mcp cursor

Manage Cursor MCP servers

### Synopsis

Manage MCP server configuration for Cursor

### Options

```
  -h, --help   help for cursor
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

* [torre mcp](torre_mcp.md)	 - MCP server management
* [torre mcp cursor disable](torre_mcp_cursor_disable.md)	 - Remove server from Cursor config
* [torre mcp cursor enable](torre_mcp_cursor_enable.md)	 - Add server to Cursor config
* [torre mcp cursor list](torre_mcp_cursor_list.md)	 - Show Cursor MCP servers

