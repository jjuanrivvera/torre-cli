## torre mcp

MCP server management

### Synopsis

Manage MCP servers for AI assistants and code editors

### Options

```
  -h, --help   help for mcp
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
* [torre mcp claude](torre_mcp_claude.md)	 - Manage Claude Desktop MCP servers
* [torre mcp cursor](torre_mcp_cursor.md)	 - Manage Cursor MCP servers
* [torre mcp start](torre_mcp_start.md)	 - Start the MCP server
* [torre mcp stream](torre_mcp_stream.md)	 - Stream the MCP server over HTTP
* [torre mcp tools](torre_mcp_tools.md)	 - Export tools as JSON
* [torre mcp vscode](torre_mcp_vscode.md)	 - Manage VSCode MCP servers

