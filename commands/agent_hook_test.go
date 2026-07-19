package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestHookScript_BashExecution exercises the generated hook with real bash. torre is
// read-only, so the hook's real job is gating the raw `api` escape hatch by HTTP method and
// blocking `alias set`; every dedicated command is a read and stays allowed.
func TestHookScript_BashExecution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash hook tests require a POSIX shell; skipping on windows")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found in PATH; skipping hook execution tests")
	}

	hookContent := hookScript(classifyAPICommands(false))
	tmpDir := t.TempDir()
	hookFile := filepath.Join(tmpDir, "torre-guard.sh")
	if err := os.WriteFile(hookFile, []byte(hookContent), 0o755); err != nil { // #nosec G306 -- hook must be executable
		t.Fatalf("write hook: %v", err)
	}

	bashPayload := func(command string) string {
		b, _ := json.Marshal(map[string]any{
			"tool_name":  "Bash",
			"tool_input": map[string]any{"command": command},
		})
		return string(b)
	}
	mcpPayload := func(toolName string) string {
		b, _ := json.Marshal(map[string]any{"tool_name": toolName, "tool_input": map[string]any{}})
		return string(b)
	}

	runHook := func(t *testing.T, payload string) string {
		t.Helper()
		cmd := exec.Command(bash, hookFile)
		cmd.Stdin = strings.NewReader(payload)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			t.Logf("hook output: %s", out.String())
			t.Fatalf("hook script exited non-zero: %v", err)
		}
		return out.String()
	}
	isDenied := func(output string) bool { return strings.Contains(output, `"permissionDecision":"deny"`) }

	cases := []struct {
		name       string
		payload    string
		wantDenied bool
	}{
		// --- alias minting ---
		{"alias_set_denied", bashPayload(`torre alias set kill "api DELETE x"`), true},
		{"quote_split_denied", bashPayload(`torre alias s""et kill "x"`), true},
		{"single_quote_split_denied", bashPayload(`torre alias s''et kill "x"`), true},
		{"backslash_denied", bashPayload(`torre alias s\et kill "x"`), true},
		// --- command position after separators ---
		{"after_semicolon_denied", bashPayload("true; torre alias set kill x"), true},
		{"after_pipe_denied", bashPayload("echo hi | torre alias set kill x"), true},
		{"after_and_denied", bashPayload("true && torre alias set kill x"), true},
		{"trailing_separator_denied", bashPayload("torre alias set kill x;true"), true},
		{"env_prefix_denied", bashPayload("env TORRE_PROFILE=w torre alias set kill x"), true},
		// --- path-invoked binaries ---
		{"relative_path_binary_denied", bashPayload("./bin/torre alias set kill x"), true},
		{"absolute_path_api_denied", bashPayload("/usr/local/bin/torre api DELETE suite/opportunities/x"), true},
		// --- raw api escape hatch (METHOD position; only GET/HEAD/OPTIONS pass) ---
		{"api_delete_denied", bashPayload("torre api DELETE suite/opportunities/x"), true},
		{"api_lowercase_delete_denied", bashPayload("torre api delete suite/opportunities/x"), true},
		{"api_post_denied", bashPayload("torre api POST opportunities/_search/ --host search -d '{}'"), true},
		{"api_put_denied", bashPayload("torre api PUT x -d '{}'"), true},
		{"api_compound_get_then_post_denied", bashPayload("torre api GET x;torre api POST y -d '{}'"), true},
		// --- raw api reads stay allowed ---
		{"api_get_allowed", bashPayload("torre api GET genome/bios/x"), false},
		{"api_get_lowercase_allowed", bashPayload("torre api get genome/bios/x"), false},
		{"api_head_allowed", bashPayload("torre api HEAD x"), false},
		{"api_get_delete_in_path_allowed", bashPayload("torre api GET suite/opportunities/deleted-role"), false},
		// --- benign reads that must stay allowed ---
		{"jobs_search_allowed", bashPayload("torre jobs search --skill golang"), false},
		{"genome_allowed", bashPayload("torre genome torrenegra"), false},
		{"search_with_delete_in_arg_allowed", bashPayload(`torre jobs search --skill "how to delete data"`), false},
		{"alias_list_allowed", bashPayload("torre alias list"), false},
		{"alias_remove_allowed", bashPayload("torre alias remove go"), false},
		{"quoted_blocked_cmd_in_arg_denied_conservatively", bashPayload(`rg "torre alias set" docs/`), true},
		{"cat_file_allowed", bashPayload("cat jobs.go"), false},
		{"other_binary_allowed", bashPayload("mytorre alias set kill x"), false},
		{"other_binary_api_allowed", bashPayload("mytorre api DELETE x"), false},
		{"opps_alias_search_allowed", bashPayload("torre opps search --skill go"), false},
		// --- MCP branch: torre has no blocked tools; read tools allowed ---
		{"mcp_jobs_search_allowed", mcpPayload("mcp__torre__torre_jobs_search"), false},
		{"mcp_genome_allowed", mcpPayload("mcp__torre__torre_genome"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := runHook(t, tc.payload)
			if denied := isDenied(output); denied != tc.wantDenied {
				t.Errorf("want denied=%v, got denied=%v\noutput: %s", tc.wantDenied, denied, output)
			}
		})
	}
}

// TestHookScript_BashExecutionNoJq exercises the no-jq fallback with a STRICT PATH so jq is
// genuinely unreachable (GOAL.md §3b #3).
func TestHookScript_BashExecutionNoJq(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash hook tests require a POSIX shell; skipping on windows")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found in PATH; skipping hook execution tests")
	}

	hookContent := hookScript(classifyAPICommands(false))
	tmpDir := t.TempDir()
	hookFile := filepath.Join(tmpDir, "torre-guard.sh")
	if err := os.WriteFile(hookFile, []byte(hookContent), 0o755); err != nil { // #nosec G306 -- hook must be executable
		t.Fatalf("write hook: %v", err)
	}

	binDir := filepath.Join(tmpDir, "nojq-bin")
	if err := os.Mkdir(binDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, tool := range []string{"cat", "tr", "grep", "sed", "printf", "env"} {
		p, lerr := exec.LookPath(tool)
		if lerr != nil {
			continue
		}
		if serr := os.Symlink(p, filepath.Join(binDir, tool)); serr != nil {
			t.Fatalf("symlink %s: %v", tool, serr)
		}
	}

	bashPayload := func(command string) string {
		b, _ := json.Marshal(map[string]any{
			"tool_name":  "Bash",
			"tool_input": map[string]any{"command": command},
		})
		return string(b)
	}

	runHookNoJq := func(t *testing.T, payload string) string {
		t.Helper()
		cmd := exec.Command(bash, hookFile)
		cmd.Stdin = strings.NewReader(payload)
		env := make([]string, 0, len(os.Environ()))
		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, "PATH=") {
				env = append(env, e)
			}
		}
		cmd.Env = append(env, "PATH="+binDir)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			t.Logf("hook output: %s", out.String())
			t.Fatalf("hook script exited non-zero: %v", err)
		}
		return out.String()
	}
	isDenied := func(output string) bool { return strings.Contains(output, `"permissionDecision":"deny"`) }

	cases := []struct {
		name       string
		payload    string
		wantDenied bool
	}{
		{"nojq_alias_set_denied", bashPayload("torre alias set kill x"), true},
		{"nojq_obfuscated_alias_set_denied", bashPayload(`torre alias s""et kill x`), true},
		{"nojq_path_binary_denied", bashPayload("./bin/torre alias set kill x"), true},
		{"nojq_api_delete_denied", bashPayload("torre api DELETE suite/opportunities/x"), true},
		{"nojq_cat_file_allowed", bashPayload("cat jobs.go"), false},
		{"nojq_jobs_search_allowed", bashPayload("torre jobs search --skill go"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := runHookNoJq(t, tc.payload)
			if denied := isDenied(output); denied != tc.wantDenied {
				t.Errorf("want denied=%v, got denied=%v\noutput: %s", tc.wantDenied, denied, output)
			}
		})
	}
}
