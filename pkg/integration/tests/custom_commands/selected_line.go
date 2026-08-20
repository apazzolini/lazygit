package custom_commands

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var SelectedLine = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Use the {{ .SelectedLine.Number }} template variable in an interactive diff",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\nfour\nfive\n")
		shell.Commit("commit")
		shell.UpdateFile("file1", "one\ntwo\nthree\nfour changed\nfive\n")
	},
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInStagingView = false
		cfg.GetUserConfig().CustomCommands = []config.CustomCommand{
			{
				Key:     config.Keybinding{"X"},
				Context: "staging",
				Command: "printf '%s:%s' '{{ .SelectedPath }}' '{{ .SelectedLine.Number }}' > location.txt",
			},
		}
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			PressEnter()

		t.Views().Staging().
			IsFocused().
			NavigateToLine(Contains("+four changed")).
			Press(config.Keybinding{"X"})

		t.FileSystem().FileContent("location.txt", Equals("file1:4"))
	},
})
