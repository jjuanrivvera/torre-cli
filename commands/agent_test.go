package commands

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEveryAPICommandIsAnnotated locks §3b hardening #4: every runnable command outside the
// local/meta groups must carry an MCP classification annotation.
func TestEveryAPICommandIsAnnotated(t *testing.T) {
	root := NewRootCmd()
	var offenders []string
	var walk func(cmd *cobra.Command, path string)
	walk = func(cmd *cobra.Command, path string) {
		for _, child := range cmd.Commands() {
			p := strings.TrimSpace(path + " " + child.Name())
			if path == "" && (slices.Contains(localGroups, child.Name()) || child.Name() == "api") {
				continue
			}
			if child.Runnable() && kindAnnotated(child.Annotations) == "" {
				offenders = append(offenders, p)
			}
			walk(child, p)
		}
	}
	walk(root, "")
	assert.Empty(t, offenders, "unannotated API commands (annotate in the builder or add to localGroups)")
}

func kindAnnotated(ann map[string]string) string {
	for _, k := range []string{annReadOnly, annOpenWorld, annDestructive} {
		if ann[k] == "true" {
			return k
		}
	}
	return ""
}

// TestClassifyAPICommands: torre is a read-only client, so every API command classifies as
// read and there are no writes/destructive commands. The raw `api` escape hatch is excluded
// from the tree classification (gated by HTTP method in the hook/renderers instead).
func TestClassifyAPICommands(t *testing.T) {
	cls := classifyAPICommands(false)
	var reads []string
	for _, c := range cls.Read {
		reads = append(reads, c.Path)
	}
	for _, want := range []string{"jobs search", "jobs get", "genome", "people search"} {
		assert.Contains(t, reads, want, "must be read-only")
	}
	assert.Empty(t, cls.Write, "torre has no write commands")
	assert.Empty(t, cls.Destructive, "torre has no destructive commands")
}

// TestAliasCrossProduct locks §3b hardening #5: cobra alias paths are enumerated.
func TestAliasCrossProduct(t *testing.T) {
	var search apiCmdInfo
	for _, c := range classifyTree() {
		if c.Path == "jobs search" {
			search = c
			break
		}
	}
	require.NotEmpty(t, search.Path, "jobs search not found")
	// "jobs" has aliases "opportunities"/"opps".
	assert.Contains(t, search.AllPaths(), "opps search")
}

// TestMCPExcludesSetupCommands locks the MCP tool surface.
func TestMCPExcludesSetupCommands(t *testing.T) {
	for _, name := range []string{"agent", "auth", "config", "alias", "init", "doctor", "completion", "version", "api", "update", "mcp"} {
		assert.Contains(t, excludedMCPPaths, name)
	}
	for _, flag := range []string{"show-token", "profile", "base-url", "search-base-url"} {
		assert.Contains(t, secretFlags, flag)
	}
	root := NewRootCmd()
	excludedSeen, includedSeen := map[string]bool{}, map[string]bool{}
	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		for _, child := range cmd.Commands() {
			if child.Runnable() {
				if mcpExcluded(child) {
					excludedSeen[child.CommandPath()] = true
				} else {
					includedSeen[child.CommandPath()] = true
				}
			}
			walk(child)
		}
	}
	walk(root)
	assert.True(t, excludedSeen["torre auth login"])
	assert.True(t, excludedSeen["torre api"])
	assert.True(t, excludedSeen["torre update"], "self-updater must not be an MCP tool")
	assert.True(t, includedSeen["torre jobs search"])
	assert.True(t, includedSeen["torre genome"])
	assert.True(t, includedSeen["torre people search"])
}

// TestGuardRenderers smoke-checks each host output for the load-bearing content.
func TestGuardRenderers(t *testing.T) {
	cls := classifyAPICommands(false)

	claude, err := renderClaudeCode(cls)
	require.NoError(t, err)
	assert.Contains(t, claude, `Bash(torre api DELETE:*)`)
	assert.Contains(t, claude, `Bash(torre api POST:*)`)
	assert.Contains(t, claude, `Bash(torre alias set:*)`)
	assert.Contains(t, claude, `Bash(torre jobs search:*)`, "reads go in the allow bucket")
	assert.Contains(t, claude, "PreToolUse")
	assert.Contains(t, claude, "blocked_cmds=(")

	codex, err := renderCodex(cls)
	require.NoError(t, err)
	assert.Contains(t, codex, `approval_policy = "on-request"`)
	assert.Contains(t, codex, `sandbox_mode = "read-only"`)

	oc, err := renderOpenCode(cls)
	require.NoError(t, err)
	assert.Contains(t, oc, `"permission"`)
	assert.Contains(t, oc, `"bash"`)
	assert.Contains(t, oc, `"torre api DELETE*": "deny"`)
	assert.Contains(t, oc, `"torre alias set*": "deny"`)
	assert.Contains(t, oc, `"torre jobs search*": "allow"`)
}

// TestGuardCommand_HostsAndWrite runs the actual cobra command per host, including --write.
func TestGuardCommand_HostsAndWrite(t *testing.T) {
	e := newEnv(t, nil)

	out, _, err := e.run("agent", "guard", "--host", "claude-code")
	require.NoError(t, err)
	assert.Contains(t, out, "blocked_cmds=(")

	out, _, err = e.run("agent", "guard", "--host", "codex")
	require.NoError(t, err)
	assert.Contains(t, out, "sandbox_mode")

	out, _, err = e.run("agent", "guard", "--host", "opencode")
	require.NoError(t, err)
	assert.Contains(t, out, "permission")

	_, _, err = e.run("agent", "guard", "--host", "nope")
	require.Error(t, err)

	dest := filepath.Join(t.TempDir(), "guard.json")
	_, _, err = e.run("agent", "guard", "--host", "opencode", "--out", dest)
	require.NoError(t, err)
	b, err := os.ReadFile(dest) // #nosec G304 -- test temp path
	require.NoError(t, err)
	assert.Contains(t, string(b), "permission")

	wd, err := os.Getwd()
	require.NoError(t, err)
	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	_, _, err = e.run("agent", "guard", "--host", "claude-code", "--write")
	require.NoError(t, err)
	hook, err := os.ReadFile(filepath.Join(tmp, ".claude", "hooks", "torre-guard.sh")) // #nosec G304 -- test temp path
	require.NoError(t, err)
	assert.Contains(t, string(hook), "blocked_cmds=(")
	_, _, err = e.run("agent", "guard", "--host", "claude-code", "--write")
	require.Error(t, err, "refuses to overwrite existing files")
}
