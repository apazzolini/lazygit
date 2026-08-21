package custom_commands

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var SelectedLineInCommitFile = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Use the {{ .SelectedLine }} and {{ .SelectedPath }} template variables in a commit file",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\n")
		shell.Commit("commit")
	},
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInStagingView = false
		cfg.GetUserConfig().CustomCommands = []config.CustomCommand{
			{
				Key:     config.Keybinding{"X"},
				Context: "patchBuilding",
				Command: "printf '%s:%s:%s-%s' '{{ .SelectedPath }}' '{{ .SelectedLine.Number }}' '{{ .SelectedLine.Range.From }}' '{{ .SelectedLine.Range.To }}' > .git/location.txt",
			},
		}
	},
	Run: func(t *TestDriver, _ config.KeybindingConfig) {
		t.Views().Commits().
			Focus().
			PressEnter()

		t.Views().CommitFiles().
			IsFocused().
			PressEnter()

		t.Views().PatchBuilding().
			IsFocused().
			NavigateToLine(Contains("+two")).
			Press(config.Keybinding{"X"})

		t.FileSystem().FileContent(".git/location.txt", Equals("file1:2:2-2"))
	},
})
