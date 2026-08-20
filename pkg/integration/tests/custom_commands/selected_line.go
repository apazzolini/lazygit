package custom_commands

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var SelectedLine = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Expose working-tree diff locations to custom commands",
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\n")
		shell.Commit("commit")
		shell.UpdateFileAndAdd("file1", "one\ntwo\nthree\nSTAGED\n")
		shell.UpdateFile("file1", "one\ntwo\nthree\nSTAGED\nfour\nfive\n")
	},
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		cfg.GetUserConfig().Git.AutoRefresh = false
		cfg.GetUserConfig().CustomCommands = []config.CustomCommand{
			{
				Key: config.Keybinding{"X"}, Context: "global",
				Command: "printf '%s' '{{ if .SelectedLine }}{{ .SelectedPath }}:{{ .SelectedLine.Number }}:{{ .SelectedLine.Range.From }}-{{ .SelectedLine.Range.To }};{{ else }}nil{{ end }}' >> .git/location.txt",
			},
			{
				Key: config.Keybinding{"Z"}, Context: "global",
				Command: "printf '%s' {{ if .SelectedDiff }}{{ .SelectedDiff | quote }}{{ else }}{{ .SelectedDiffError | quote }}{{ end }} > .git/diff-note.txt",
			},
		}
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().IsFocused().Press(config.Keybinding{"X"}).Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/location.txt", Equals("nil"))
		t.FileSystem().FileContent(".git/diff-note.txt", Contains("Select lines in a working-tree or individual commit diff first."))
		t.Views().Files().Press(keys.Universal.FocusMainView)

		t.Views().Main().IsFocused().NavigateToLine(Contains("+four")).
			Press(keys.Universal.RangeSelectDown).
			SelectedLines(Contains("+four"), Contains("+five")).
			Tap(func() { t.Shell().UpdateFile("file1", "one\ntwo\nthree\nSTAGED\ndifferent\n") }).
			Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/diff-note.txt", Equals("file1:5-6"))
		t.FileSystem().FileContent(".git/diff-note.txt", DoesNotContain("different"))
		t.Shell().UpdateFile("file1", "one\ntwo\nthree\nSTAGED\nfour\nfive\n")
		t.Views().Files().Focus().Press(keys.Universal.FocusMainView)
		t.Views().Main().NavigateToLine(Contains("+four")).Press(keys.Universal.RangeSelectDown).
			Press(config.Keybinding{"X"}).Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/location.txt", Equals("nilfile1:6:5-6;"))
		t.FileSystem().FileContent(".git/diff-note.txt", Equals("file1:5-6"))

		t.Views().Main().Press(keys.Universal.PrevItem).Press(keys.Universal.NextItem).
			Press(keys.Universal.RangeSelectUp).
			SelectedLines(Contains("+four"), Contains("+five")).
			Press(config.Keybinding{"X"}).Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/location.txt", Equals("nilfile1:6:5-6;file1:5:5-6;"))
		t.FileSystem().FileContent(".git/diff-note.txt", Equals("file1:5-6"))

		t.Views().Main().Press(keys.Universal.TogglePanel)
		t.Views().Secondary().IsFocused().NavigateToLine(Contains("+STAGED")).
			Press(config.Keybinding{"X"}).Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/location.txt", Equals("nilfile1:6:5-6;file1:5:5-6;file1:4:4-4;"))
		t.FileSystem().FileContent(".git/diff-note.txt", Equals("index:file1:4"))
	},
})
