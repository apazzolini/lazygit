package custom_commands

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var SelectedDiff = NewIntegrationTest(NewIntegrationTestArgs{
	Description: "Recover both sides of a rendered selection and reject unsupported diff sources",
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\n")
		shell.Commit("initial")
		shell.UpdateFile("file1", "one\nstashed\nthree\n")
		shell.Stash("stash")
		shell.CreateFileAndAdd("file2", "second commit\n")
		shell.Commit("second")
		shell.UpdateFile("file1", "one\nTWO\nthree\n")
	},
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		cfg.GetUserConfig().Git.AutoRefresh = false
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{
			{Name: "columns", ColorArg: "never", Command: `printf '\033]1717;1\007'; ` +
				`printf '\033]1717;1;f;;;file1\007file1\n'; ` +
				`printf '\033]1717;1;h;1;;file1\007@@\n'; ` +
				`printf '\033]1717;1;c;1;;file1\007one    one\n'; ` +
				`printf '\033]1717;1;d;2;2;file1\007two    \033]1717;1;a;2;;file1\007TWO\n'; ` +
				`printf '\033]1717;1;c;3;;file1\007three  three\n'; cat >/dev/null`},
			{Name: "unified", Command: `printf '\033]1717;1\007'; cat`},
		}
		cfg.GetUserConfig().CustomCommands = []config.CustomCommand{
			{
				Key: config.Keybinding{"Y"}, Context: "normal, normalSecondary",
				Command: "{{ if .SelectedDiff }}cat > .git/sent-note.txt{{ else }}printf '%s' {{ .SelectedDiffError | quote }} >&2; exit 1{{ end }}",
				Stdin:   "{{ .SelectedDiff }}\n{{ .Form.Note }}",
				Prompts: []config.CustomCommandPrompt{{Type: "input", Key: "Note", Title: "Note", Condition: "{{ if .SelectedDiff }}true{{ end }}"}},
			},
			{
				Key: config.Keybinding{"Z"}, Context: "normal, normalSecondary",
				Command: "printf '%s' {{ if .SelectedDiff }}{{ .SelectedDiff | quote }}{{ else }}{{ .SelectedDiffError | quote }}{{ end }} > .git/diff-note.txt",
			},
		}
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().IsFocused().Press(keys.Universal.FocusMainView)
		t.Views().Main().IsFocused().NavigateToLine(Contains("two    TWO")).
			Tap(func() { t.Shell().UpdateFile("file1", "one\ndifferent\nthree\n") }).
			Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/diff-note.txt", Equals("file1:2\n```text\ntwo    TWO\n```"))
		t.FileSystem().FileContent(".git/diff-note.txt", DoesNotContain("different"))
		t.Shell().UpdateFile("file1", "one\nTWO\nthree\n")
		t.Views().Main().Press(config.Keybinding{"Y"})
		t.ExpectPopup().Prompt().Title(Equals("Note")).Type("Why replace this?").Confirm()
		t.FileSystem().FileContent(".git/sent-note.txt", Equals("file1:2\n```text\ntwo    TWO\n```\nWhy replace this?"))

		t.Views().Main().Press(keys.Universal.CycleDiffRenderers).
			Tap(func() { t.ExpectToast(Equals("Diff renderer: unified (2 of 2)")) }).
			NavigateToLine(Contains("+TWO")).Press(config.Keybinding{"Y"})
		t.ExpectPopup().Prompt().Title(Equals("Note")).Type("Why change this?").Confirm()
		t.FileSystem().FileContent(".git/sent-note.txt", Equals("file1:2\nWhy change this?"))
		t.Views().Stash().Focus().Press(keys.Universal.FocusMainView)
		t.Views().Main().IsFocused().NavigateToLine(Contains("+stashed")).Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/diff-note.txt", Contains("not stashes or commit ranges"))

		t.Views().Commits().Focus().Press(keys.Universal.RangeSelectDown).Press(keys.Universal.FocusMainView)
		t.Views().Main().IsFocused().NavigateToLine(Contains("+second commit")).Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/diff-note.txt", Contains("not stashes or commit ranges"))

		t.Views().Commits().Focus().NavigateToLine(Contains("second")).
			Press(keys.Universal.DiffingMenu).
			Tap(func() { t.ExpectPopup().Menu().Title(Equals("Diffing")).Select(MatchesRegexp(`Diff \w+`)).Confirm() }).
			NavigateToLine(Contains("initial")).Press(keys.Universal.FocusMainView)
		t.Views().Main().IsFocused().Content(Contains("-second commit")).Press(config.Keybinding{"Z"})
		t.FileSystem().FileContent(".git/diff-note.txt", Contains("Exit revision comparison mode before sending a note."))
		t.Views().Main().Press(config.Keybinding{"Y"})
		t.ExpectPopup().Alert().Title(Equals("Error")).Content(Contains("Exit revision comparison mode before sending a note.")).Confirm()
	},
})
