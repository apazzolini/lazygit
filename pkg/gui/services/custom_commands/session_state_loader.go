package custom_commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/controllers/helpers"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/samber/lo"
)

// loads the session state at the time that a custom command is invoked, for use
// in the custom command's template strings
type SessionStateLoader struct {
	c          *helpers.HelperCommon
	refsHelper *helpers.RefsHelper
}

func NewSessionStateLoader(c *helpers.HelperCommon, refsHelper *helpers.RefsHelper) *SessionStateLoader {
	return &SessionStateLoader{
		c:          c,
		refsHelper: refsHelper,
	}
}

func commitShimFromModelCommit(commit *models.Commit) *Commit {
	if commit == nil {
		return nil
	}

	return &Commit{
		Hash:          commit.Hash(),
		Sha:           commit.Hash(),
		Name:          commit.Name,
		Status:        commit.Status,
		Action:        commit.Action,
		Tags:          commit.Tags,
		ExtraInfo:     commit.ExtraInfo,
		AuthorName:    commit.AuthorName,
		AuthorEmail:   commit.AuthorEmail,
		UnixTimestamp: commit.UnixTimestamp,
		Divergence:    commit.Divergence,
		Parents:       commit.Parents(),
	}
}

func fileShimFromModelFile(file *models.File) *File {
	if file == nil {
		return nil
	}

	return &File{
		Name:                    file.Path,
		PreviousName:            file.PreviousPath,
		HasStagedChanges:        file.HasStagedChanges,
		HasUnstagedChanges:      file.HasUnstagedChanges,
		Tracked:                 file.Tracked,
		Added:                   file.Added,
		Deleted:                 file.Deleted,
		HasMergeConflicts:       file.HasMergeConflicts,
		HasInlineMergeConflicts: file.HasInlineMergeConflicts,
		DisplayString:           file.DisplayString,
		ShortStatus:             file.ShortStatus,
		IsWorktree:              file.IsWorktree,
	}
}

func submoduleShimFromModelSubmodule(submodule *models.SubmoduleConfig) *Submodule {
	if submodule == nil {
		return nil
	}

	return &Submodule{
		Name: submodule.Name,
		Path: submodule.Path,
		Url:  submodule.Url,
	}
}

func branchShimFromModelBranch(branch *models.Branch) *Branch {
	if branch == nil {
		return nil
	}

	return &Branch{
		Name:           branch.Name,
		DisplayName:    branch.DisplayName,
		Recency:        branch.Recency,
		Pushables:      branch.AheadForPull,
		Pullables:      branch.BehindForPull,
		AheadForPull:   branch.AheadForPull,
		BehindForPull:  branch.BehindForPull,
		AheadForPush:   branch.AheadForPush,
		BehindForPush:  branch.BehindForPush,
		UpstreamGone:   branch.UpstreamGone,
		Head:           branch.Head,
		DetachedHead:   branch.DetachedHead,
		UpstreamRemote: branch.UpstreamRemote,
		UpstreamBranch: branch.UpstreamBranch,
		Subject:        branch.Subject,
		CommitHash:     branch.CommitHash,
	}
}

func remoteBranchShimFromModelRemoteBranch(remoteBranch *models.RemoteBranch) *RemoteBranch {
	if remoteBranch == nil {
		return nil
	}

	return &RemoteBranch{
		Name:       remoteBranch.Name,
		RemoteName: remoteBranch.RemoteName,
	}
}

func remoteShimFromModelRemote(remote *models.Remote) *Remote {
	if remote == nil {
		return nil
	}

	return &Remote{
		Name:     remote.Name,
		Urls:     remote.Urls,
		PushUrls: remote.PushUrls,
		Branches: lo.Map(remote.Branches, func(branch *models.RemoteBranch, _ int) *RemoteBranch {
			return remoteBranchShimFromModelRemoteBranch(branch)
		}),
	}
}

func tagShimFromModelRemote(tag *models.Tag) *Tag {
	if tag == nil {
		return nil
	}

	return &Tag{
		Name:    tag.Name,
		Message: tag.Message,
	}
}

func stashEntryShimFromModelRemote(stashEntry *models.StashEntry) *StashEntry {
	if stashEntry == nil {
		return nil
	}

	return &StashEntry{
		Index:   stashEntry.Index,
		Recency: stashEntry.Recency,
		Name:    stashEntry.Name,
	}
}

func commitFileShimFromModelRemote(commitFile *models.CommitFile) *CommitFile {
	if commitFile == nil {
		return nil
	}

	return &CommitFile{
		Name:         commitFile.Path,
		ChangeStatus: commitFile.ChangeStatus,
	}
}

func worktreeShimFromModelRemote(worktree *models.Worktree) *Worktree {
	if worktree == nil {
		return nil
	}

	return &Worktree{
		IsMain:        worktree.IsMain,
		IsCurrent:     worktree.IsCurrent,
		Path:          worktree.Path,
		IsPathMissing: worktree.IsPathMissing,
		GitDir:        worktree.GitDir,
		Branch:        worktree.Branch,
		Name:          worktree.Name,
	}
}

type CommitRange struct {
	From string
	To   string
}

func makeCommitRange(commits []*models.Commit, _ int, _ int) *CommitRange {
	if len(commits) == 0 {
		return nil
	}

	return &CommitRange{
		From: commits[len(commits)-1].Hash(),
		To:   commits[0].Hash(),
	}
}

// SessionState captures the current state of the application for use in custom commands
type SessionState struct {
	SelectedLocalCommit    *Commit // deprecated, use SelectedCommit
	SelectedReflogCommit   *Commit // deprecated, use SelectedCommit
	SelectedSubCommit      *Commit // deprecated, use SelectedCommit
	SelectedCommit         *Commit
	SelectedCommitRange    *CommitRange
	SelectedFile           *File
	SelectedLine           *Line
	SelectedDiff           string
	SelectedDiffError      string
	SelectedSubmodule      *Submodule
	SelectedPath           string
	SelectedLocalBranch    *Branch
	SelectedRemoteBranch   *RemoteBranch
	SelectedRemote         *Remote
	SelectedTag            *Tag
	SelectedStashEntry     *StashEntry
	SelectedCommitFile     *CommitFile
	SelectedCommitFilePath string
	SelectedWorktree       *Worktree
	CheckedOutBranch       *Branch
}

func (self *SessionStateLoader) call() *SessionState {
	selectedLocalCommit := commitShimFromModelCommit(self.c.Contexts().LocalCommits.GetSelected())
	selectedLocalCommitRange := makeCommitRange(self.c.Contexts().LocalCommits.GetSelectedItems())
	selectedReflogCommit := commitShimFromModelCommit(self.c.Contexts().ReflogCommits.GetSelected())
	selectedReflogCommitRange := makeCommitRange(self.c.Contexts().ReflogCommits.GetSelectedItems())
	selectedSubCommit := commitShimFromModelCommit(self.c.Contexts().SubCommits.GetSelected())
	selectedSubCommitRange := makeCommitRange(self.c.Contexts().SubCommits.GetSelectedItems())

	selectedCommit := selectedLocalCommit
	selectedCommitRange := selectedLocalCommitRange
	if self.c.Context().IsCurrentOrParent(self.c.Contexts().ReflogCommits) {
		selectedCommit = selectedReflogCommit
		selectedCommitRange = selectedReflogCommitRange
	} else if self.c.Context().IsCurrentOrParent(self.c.Contexts().SubCommits) {
		selectedCommit = selectedSubCommit
		selectedCommitRange = selectedSubCommitRange
	}

	selectedPath := self.c.Contexts().Files.GetSelectedPath()
	selectedCommitFilePath := self.c.Contexts().CommitFiles.GetSelectedPath()

	if self.c.Context().IsCurrent(self.c.Contexts().CommitFiles) {
		selectedPath = selectedCommitFilePath
	}

	var selectedLine *Line
	selectedDiff := ""
	selectedDiffError := self.c.Tr.DiffNoteSelectLines
	for _, diffContext := range []*context.MainContext{
		self.c.Contexts().Normal,
		self.c.Contexts().NormalSecondary,
	} {
		if !self.c.Context().IsCurrent(diffContext) {
			continue
		}

		if self.c.Modes().Diffing.Active() {
			selectedDiffError = self.c.Tr.DiffNoteComparison
		}
		view := diffContext.GetView()
		diffLineHelper := helpers.NewDiffLineHelper(self.c)
		cursor, cursorOk := diffLineHelper.GetDiffLineInfo(view, view.SelectedLineIdx())
		if !cursorOk {
			break
		}
		if path, err := filepath.Rel(self.c.Git().RepoPaths.WorktreePath(), cursor.Path); err == nil {
			selectedPath = filepath.ToSlash(path)
		}
		if !view.Highlight {
			break
		}

		from, to := view.SelectedLineRange()
		first, firstOk := diffLineHelper.GetDiffLineInfo(view, from)
		last, lastOk := diffLineHelper.GetDiffLineInfo(view, to)
		bufferFrom, bufferTo, rangeOk := view.SelectedBufferLineRange()
		if !firstOk || !lastOk || !rangeOk {
			break
		}
		selectedInfos := diffLineHelper.DiffLinesInBufferRange(view, bufferFrom, bufferTo)
		singleFile := true
		for _, info := range selectedInfos {
			if info.Path != cursor.Path {
				singleFile = false
				break
			}
		}
		if !singleFile {
			selectedDiffError = self.c.Tr.DiffNoteOneFile
			break
		}

		for _, info := range []*types.DiffLineInfo{&cursor, &first, &last} {
			if info.Type == types.DiffLineFileHeader {
				info.NewLine = 1
			}
		}
		selectedLine = &Line{
			Number: cursor.NewLine,
			Range: &LineRange{
				From: first.NewLine,
				To:   last.NewLine,
			},
		}
		if diffLineHelper.ShowsCustomPatch(view) {
			selectedDiffError = self.c.Tr.DiffNoteCustomPatch
			break
		}
		if self.c.Modes().Diffing.Active() {
			selectedDiffError = self.c.Tr.DiffNoteComparison
			break
		}
		if !self.c.Context().IsInStack(diffContext) {
			break
		}
		source := self.c.Context().NextInStack(diffContext)
		var commit *models.Commit
		diffKind := ""
		switch source := source.(type) {
		case *context.WorkingTreeContext:
			diffKind = "unstaged"
			if diffContext == self.c.Contexts().NormalSecondary {
				diffKind = "index"
			}
		case *context.LocalCommitsContext:
			commits, _, _ := source.GetSelectedItems()
			if len(commits) == 1 {
				commit = source.GetSelected()
			}
		case *context.SubCommitsContext:
			commits, _, _ := source.GetSelectedItems()
			if len(commits) == 1 {
				commit = source.GetSelected()
			}
		case *context.ReflogCommitsContext:
			commits, _, _ := source.GetSelectedItems()
			if len(commits) == 1 {
				commit = source.GetSelected()
			}
		case *context.CommitFilesContext:
			if source.GetRefRange() == nil {
				commit, _ = source.GetRef().(*models.Commit)
			}
		}
		if commit != nil {
			diffKind = commit.ShortHash()
		}
		if diffKind == "" {
			selectedDiffError = self.c.Tr.DiffNoteSource
			break
		}
		displayedLines := view.BufferLines()
		if bufferFrom < 0 || bufferTo >= len(displayedLines) || len(selectedInfos) == 0 {
			selectedDiffError = self.c.Tr.DiffNoteUnresolved
			break
		}
		newFrom, newTo := -1, -1
		oldFrom, oldTo := -1, -1
		includeSnippet := false
		for _, info := range selectedInfos {
			switch info.Type {
			case types.DiffLineDeleted:
				includeSnippet = true
				if oldFrom == -1 || info.OldLine < oldFrom {
					oldFrom = info.OldLine
				}
				oldTo = max(oldTo, info.OldLine)
			case types.DiffLineAdded, types.DiffLineContext:
				if newFrom == -1 || info.NewLine < newFrom {
					newFrom = info.NewLine
				}
				newTo = max(newTo, info.NewLine)
			default:
				includeSnippet = true
			}
		}
		var location strings.Builder
		if diffKind != "unstaged" {
			fmt.Fprintf(&location, "%s:", diffKind)
		}
		fmt.Fprintf(&location, "%s:", selectedPath)
		deletedOnly := newFrom == -1 && oldFrom != -1
		if deletedOnly {
			location.WriteString("old ")
			newFrom, newTo = oldFrom, oldTo
		} else if newFrom == -1 {
			newFrom, newTo = selectedLine.Range.From, selectedLine.Range.To
		}
		fmt.Fprintf(&location, "%d", newFrom)
		if newFrom != newTo {
			fmt.Fprintf(&location, "-%d", newTo)
		}
		if deletedOnly {
			location.WriteString(" (deleted)")
		}
		if includeSnippet {
			fmt.Fprintf(&location, "\n```text\n%s\n```",
				strings.Join(displayedLines[bufferFrom:bufferTo+1], "\n"))
		}
		selectedDiff = location.String()
		selectedDiffError = ""
		break
	}

	return &SessionState{
		SelectedFile:           fileShimFromModelFile(self.c.Contexts().Files.GetSelectedFile()),
		SelectedLine:           selectedLine,
		SelectedDiff:           selectedDiff,
		SelectedDiffError:      selectedDiffError,
		SelectedSubmodule:      submoduleShimFromModelSubmodule(self.c.Contexts().Submodules.GetSelected()),
		SelectedPath:           selectedPath,
		SelectedLocalCommit:    selectedLocalCommit,
		SelectedReflogCommit:   selectedReflogCommit,
		SelectedSubCommit:      selectedSubCommit,
		SelectedCommit:         selectedCommit,
		SelectedCommitRange:    selectedCommitRange,
		SelectedLocalBranch:    branchShimFromModelBranch(self.c.Contexts().Branches.GetSelected()),
		SelectedRemoteBranch:   remoteBranchShimFromModelRemoteBranch(self.c.Contexts().RemoteBranches.GetSelected()),
		SelectedRemote:         remoteShimFromModelRemote(self.c.Contexts().Remotes.GetSelected()),
		SelectedTag:            tagShimFromModelRemote(self.c.Contexts().Tags.GetSelected()),
		SelectedStashEntry:     stashEntryShimFromModelRemote(self.c.Contexts().Stash.GetSelected()),
		SelectedCommitFile:     commitFileShimFromModelRemote(self.c.Contexts().CommitFiles.GetSelectedFile()),
		SelectedCommitFilePath: selectedCommitFilePath,
		SelectedWorktree:       worktreeShimFromModelRemote(self.c.Contexts().Worktrees.GetSelected()),
		CheckedOutBranch:       branchShimFromModelBranch(self.refsHelper.GetCheckedOutRef()),
	}
}
