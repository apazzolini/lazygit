package custom_commands

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var SelectedLine = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Use the {{ .SelectedLine }} template variable in an interactive diff",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\n")
		shell.Commit("commit")
		shell.UpdateFile("file1", "one\ntwo\nthree\nfour\nfive\n")
	},
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInStagingView = false
		cfg.GetUserConfig().CustomCommands = []config.CustomCommand{
			{
				Key:     config.Keybinding{"X"},
				Context: "staging",
				Command: "printf '%s:%s:%s-%s' '{{ .SelectedPath }}' '{{ .SelectedLine.Number }}' '{{ .SelectedLine.Range.From }}' '{{ .SelectedLine.Range.To }}' > location.txt",
			},
		}
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			PressEnter()

		t.Views().Staging().
			IsFocused().
			NavigateToLine(Contains("+four")).
			Press(keys.Universal.RangeSelectDown).
			SelectedLines(
				Contains("+four"),
				Contains("+five"),
			).
			Press(config.Keybinding{"X"})

		t.FileSystem().FileContent("location.txt", Equals("file1:5:4-5"))
	},
})
