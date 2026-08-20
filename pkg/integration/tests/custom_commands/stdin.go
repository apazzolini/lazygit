package custom_commands

import (
	"strings"

	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var Stdin = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Send large templated input directly to a custom command after prompting",
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\n")
		shell.Commit("initial")
	},
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().CustomCommands = []config.CustomCommand{
			{
				Key: config.Keybinding{"Y"}, Context: "files",
				Command: "cat > .git/input.txt",
				Stdin:   strings.Repeat("large payload\n", 12000) + "{{ .SelectedPath }}\n{{ .Form.Note }}",
				Prompts: []config.CustomCommandPrompt{{Type: "input", Key: "Note", Title: "Note"}},
			},
			{
				Key: config.Keybinding{"Z"}, Context: "files",
				Command: "cat > .git/log-input.txt", Stdin: strings.Repeat("large payload\n", 12000), Output: "log",
			},
			{
				Key: config.Keybinding{"X"}, Context: "files",
				Command: "cat > .git/terminal-input.txt", Stdin: "terminal input", Output: "terminal",
			},
			{
				Key: config.Keybinding{"P"}, Context: "files",
				Command: "cat", Stdin: "popup input", Output: "popup", OutputTitle: "Input",
			},
		}
	},
	Run: func(t *TestDriver, _ config.KeybindingConfig) {
		t.Views().Files().IsFocused().Press(config.Keybinding{"Y"})
		t.ExpectPopup().Prompt().Title(Equals("Note")).Type("Why 'this' $(not-a-command)?").Confirm()
		t.FileSystem().FileContent(".git/input.txt", Equals(strings.Repeat("large payload\n", 12000)+"\nWhy 'this' $(not-a-command)?"))
		t.Views().Files().Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/log-input.txt", Equals(strings.Repeat("large payload\n", 12000)))
		t.Views().Files().Press(config.Keybinding{"X"})
		t.FileSystem().FileContent(".git/terminal-input.txt", Equals("terminal input"))
		t.Views().Files().Press(config.Keybinding{"P"})
		t.ExpectPopup().Alert().Title(Equals("Input")).Content(Equals("popup input")).Confirm()
	},
})
