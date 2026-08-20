package custom_commands

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var SelectedLineInCommitFile = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Expose single-file commit diff locations and reject unsupported selections",
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\nold\nthree\n")
		shell.Commit("initial")
		shell.UpdateFileAndAdd("file1", "one\ntwo\nthree\n")
		shell.CreateFileAndAdd("file2", "other file\n")
		shell.Commit("commit")
	},
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		cfg.GetUserConfig().CustomCommands = []config.CustomCommand{
			{
				Key: config.Keybinding{"X"}, Context: "normal, normalSecondary",
				Command: "printf '%s' '{{ if .SelectedLine }}{{ .SelectedPath }}:{{ .SelectedLine.Number }}:{{ .SelectedLine.Range.From }}-{{ .SelectedLine.Range.To }};{{ else }}nil{{ end }}' >> .git/location.txt",
			},
			{
				Key: config.Keybinding{"Z"}, Context: "normal, normalSecondary",
				Command: "printf '%s' {{ if .SelectedDiff }}{{ .SelectedDiff | quote }}{{ else }}{{ .SelectedDiffError | quote }}{{ end }} > .git/diff-note.txt",
			},
		}
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Commits().Focus().Press(keys.Universal.FocusMainView)
		t.Views().Main().IsFocused().NavigateToLine(Contains("-old")).Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/diff-note.txt", Equals(t.Git().GetCommitHash("HEAD")[:8]+":file1:old 2 (deleted)\n```text\n-old\n```"))

		t.Views().Commits().Focus().PressEnter()
		t.Views().CommitFiles().IsFocused().Press(keys.Universal.FocusMainView)
		t.Views().Main().IsFocused().NavigateToLine(Contains("+two")).
			Press(config.Keybinding{"X"}).Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/location.txt", Equals("file1:2:2-2;"))
		t.FileSystem().FileContent(".git/diff-note.txt", Equals(t.Git().GetCommitHash("HEAD")[:8]+":file1:2"))

		t.Views().Main().Press(keys.Universal.ToggleRangeSelect).
			NavigateToLine(Contains("+other file")).
			Press(config.Keybinding{"X"}).Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/location.txt", Equals("file1:2:2-2;nil"))
		t.FileSystem().FileContent(".git/diff-note.txt", Contains("Select lines from only one file"))

		t.Views().Main().Press(keys.Universal.PrevItem).
			NavigateToLine(Contains("+two")).PressPrimaryAction().Press(keys.Universal.TogglePanel)
		t.Views().Secondary().IsFocused().NavigateToLine(Contains("+two")).Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/diff-note.txt", Contains("Notes from custom-patch previews are not supported."))
	},
})
