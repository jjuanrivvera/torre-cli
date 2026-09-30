# torre demo — rendered with VHS (https://github.com/charmbracelet/vhs).
#
# NO ACCOUNT, NO SIGNUP, NO KEYRING, NO NETWORK beyond loopback. All data is
# invented by tools/demoserver; tables, JSON, --jq and command execution are real.
# Configuration is isolated from the recorder's own profiles and credentials.
#
# Regenerate:
#   go build -o bin/torre ./cmd/torre && go build -o bin/demoserver ./tools/demoserver && vhs .demo/demo.tape

Output assets/demo.gif

Set Shell bash
Set FontSize 15
Set Width 1200
Set Height 760
Set Padding 18
Set Theme "Dracula"
Set TypingSpeed 28ms

Hide
Type "export PATH=$PWD/bin:$PATH XDG_CONFIG_HOME=$(mktemp -d) TORRE_TOKEN=demo-token-no-account"
Enter
Type "demoserver 127.0.0.1:8644 >/dev/null 2>&1 & demo_pid=$!; trap 'kill $demo_pid 2>/dev/null; rm -rf $XDG_CONFIG_HOME' EXIT; sleep 1"
Enter
Type "alias torre='torre --base-url http://127.0.0.1:8644 --search-base-url http://127.0.0.1:8644 --profile demo'; clear"
Enter
Show

Sleep 500ms

Type "torre jobs search --skill golang --remote-anywhere --limit 2 -o table"
Sleep 250ms
Enter
Sleep 2400ms

Type "torre jobs get demo-go-01 -o json"
Sleep 250ms
Enter
Sleep 2400ms

Type "clear"
Enter
Sleep 300ms

Type "torre people search --skill golang --remote --limit 2 -o table"
Sleep 250ms
Enter
Sleep 2400ms

Type "clear"
Enter
Sleep 300ms

Type "torre genome demo_candidate_a -o json --jq '[.strengths[].name]'"
Sleep 250ms
Enter
Sleep 2400ms
