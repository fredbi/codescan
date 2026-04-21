// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package grammar

import (
	"go/ast"
	"go/token"
	"strings"
)

// Line is one preprocessed comment line ready for the lexer.
//
// Text has the Go comment markers (// /* */) stripped, along with
// leading continuation decorations common in godoc comments (spaces,
// tabs, asterisks, slashes, dashes, optional markdown table pipe).
// Internal content and embedded whitespace are preserved — fence-body
// indentation handling lives at the lexer layer where fence state is
// tracked.
//
// Pos is the position of Text's first character in the source file.
// For continuation lines inside a /* … */ block, the column is
// approximated to 1 — exact column reconstruction would require
// re-tokenising the comment body and is deferred until LSP needs it.
type Line struct {
	Text string
	Pos  token.Position
}

// Preprocess turns a comment group into a position-tagged []Line.
//
// Nil CommentGroup or FileSet returns nil. The function is pure: it
// makes no syscalls, allocates a slice proportional to the number of
// physical lines, and is safe for concurrent use.
//
// See architecture §3.1 (stage diagram) and tasks P1.1.
func Preprocess(cg *ast.CommentGroup, fset *token.FileSet) []Line {
	if cg == nil || fset == nil {
		return nil
	}
	var out []Line
	for _, c := range cg.List {
		out = append(out, stripComment(c.Text, fset.Position(c.Slash))...)
	}
	return out
}

// stripComment returns one Line per physical source line of a single
// *ast.Comment. It handles both the `//` line-comment form and the
// `/* … */` block form, including multi-line blocks.
func stripComment(raw string, basePos token.Position) []Line {
	switch {
	case strings.HasPrefix(raw, "//"):
		text := trimContentPrefix(strings.TrimPrefix(raw, "//"))
		return []Line{{Text: text, Pos: basePos}}
	case strings.HasPrefix(raw, "/*"):
		body := strings.TrimSuffix(strings.TrimPrefix(raw, "/*"), "*/")
		rawLines := strings.Split(body, "\n")
		out := make([]Line, 0, len(rawLines))
		for i, r := range rawLines {
			pos := basePos
			pos.Line += i
			if i > 0 {
				pos.Column = 1
			}
			out = append(out, Line{Text: trimContentPrefix(r), Pos: pos})
		}
		return out
	default:
		// Not a valid Go comment; preserve input defensively so
		// downstream layers can surface a diagnostic rather than
		// silently lose data.
		return []Line{{Text: raw, Pos: basePos}}
	}
}

// trimContentPrefix removes the leading godoc-style decoration that
// precedes real content on a comment line:
//   - whitespace (space, tab)
//   - continuation slashes and asterisks (“//“, “ * “, “ *  “)
//   - dashes (“ -- “)
//   - an optional single markdown table pipe “|“
//
// The set mirrors the v1 parser's rxUncommentHeaders so migrated
// fixtures match byte-for-byte at the parse-output level (pre-P5
// parity harness).
func trimContentPrefix(s string) string {
	s = strings.TrimLeft(s, " \t*/-")
	s = strings.TrimPrefix(s, "|")
	return strings.TrimLeft(s, " \t")
}
