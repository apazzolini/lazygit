package custom_commands

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var SelectedDiffRename = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Capture deleted and added lines of a renamed file as displayed",
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("old.txt", "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\n")
		shell.Commit("initial")
		shell.RenameFileInGit("old.txt", "renamed.txt")
		shell.UpdateFileAndAdd("renamed.txt", "one\nTWO\nthree\nfour\nfive\nsix\nseven\neight\n")
		shell.Commit("rename")
	},
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		cfg.GetUserConfig().CustomCommands = []config.CustomCommand{
			{Key: config.Keybinding{"Z"}, Context: "normal", Command: "cat > .git/diff-note.txt", Stdin: "{{ .SelectedDiff }}"},
		}
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Commits().Focus().Press(keys.Universal.FocusMainView)
		t.Views().Main().IsFocused().Content(Contains("rename from old.txt")).
			NavigateToLine(Contains("-two")).Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/diff-note.txt", Equals(t.Git().GetCommitHash("HEAD")[:8]+":renamed.txt:old 2 (deleted)\n```text\n-two\n```"))
		t.FileSystem().FileContent(".git/diff-note.txt", DoesNotContain("new file mode"))
		t.Views().Main().NavigateToLine(Contains("+TWO")).Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/diff-note.txt", Equals(t.Git().GetCommitHash("HEAD")[:8]+":renamed.txt:2"))
	},
})
