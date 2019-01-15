package patchfile

import (
	"errors"
)

// ErrNoDifference is returned if the two directory trees Apply has been
// called have the same tree hash.
var ErrNoDifference = errors.New("patchfile: directory trees do not differ")

// ErrHeaderFieldsNum is returned if the header does not have 4 space separated fields.
var ErrHeaderFieldsNum = errors.New("patchfile: header does not have 4 space separated fields")

// ErrHeaderFieldsText is returned if the header does not start with 'codechain patchfile version'.
var ErrHeaderFieldsText = errors.New("patchfile: header does not start with 'codechain patchfile version'")

// ErrHeaderVersion is returned if if the patchfile version is not supported.
var ErrHeaderVersion = errors.New("patchfile: version not supported")

// ErrTreeHashFieldsNum is returned if the tree hash does not have 2 space separated fields.
var ErrTreeHashFieldsNum = errors.New("patchfile: treehash does not have 2 space separated fields")

// ErrTreeHashFieldsText is returned if the tree hash does not start with 'codechain patchfile version'.
var ErrTreeHashFieldsText = errors.New("patchfile: treehash line does not start with 'treehash'")

// ErrTreeHashStartMismatch is returned if directory tree hash does not match the one given in the treehash start line.
var ErrTreeHashStartMismatch = errors.New("patchfile: directory tree hash does not match treehash start line")

// ErrTreeHashFinishMismatch is returned if directory tree hash does not match the one given in the treehash finish line.
var ErrTreeHashFinishMismatch = errors.New("patchfile: directory tree hash does not match treehash finish line")

// ErrFileHashMismatchBefore is returned if file hash does not match before apply.
var ErrFileHashMismatchBefore = errors.New("patchfile: file hash does not match before apply")

// ErrFileHashMismatchAfter is returned if file hash does not match after apply.
var ErrFileHashMismatchAfter = errors.New("patchfile: file hash does not match after apply")

// ErrFileFieldsNum is returned if a file diff line does not have 4 space separated fields.
var ErrFileFieldsNum = errors.New("patchfile: file diff line does not have 4 space separated fields")

// ErrFileField0 is returned if the file diff line does not start with '-' or '+'.
var ErrFileField0 = errors.New("patchfile: file diff line does not start with '-' or '+'")

// ErrFileField1 is returned if the file diff line does have mode 'f' or 'x'.
var ErrFileField1 = errors.New("patchfile: file diff line does not have mode 'f' or 'x'")

// ErrAddTargetFileExists is returned if an add target file exists already.
var ErrAddTargetFileExists = errors.New("patchfile: add target file exists already")

// ErrMoveTargetFileExists is returned if a move target file exists already.
var ErrMoveTargetFileExists = errors.New("patchfile: move target file exists already")

// ErrDiffLinesParse is returned if the number of diff lines cannot be parsed.
var ErrDiffLinesParse = errors.New("patchfile: cannot parse number of diff lines")

// ErrDiffLinesNonPositive is returned if the number of diff lines is non-positive.
var ErrDiffLinesNonPositive = errors.New("patchfile: number of lines < 1")

// ErrDiffLinesNegative is returned if the number of diff lines is negative.
var ErrDiffLinesNegative = errors.New("patchfile: number of lines < 0")

// ErrDiffModeUnknown if returned if the diff mode is unknown.
var ErrDiffModeUnknown = errors.New("patchfile: unknown diff modes")

// ErrNotTerminal is returned if more input is read after terminal state.
var ErrNotTerminal = errors.New("patchfile: more input read after terminal state")

// ErrPrematureDiffEnd is returned if a diff ends prematurely.
var ErrPrematureDiffEnd = errors.New("patchfile: diff ends prematurely")

// ErrPrematurePatchfileEnd is returned if a patchfile ends prematurely.
var ErrPrematurePatchfileEnd = errors.New("patchfile: ends prematurely")

// ErrDiffNotClean is returned if no clean diff could be computed.
var ErrDiffNotClean = errors.New("patchfile: computed diff is not clean (use patchfile version 2)")
