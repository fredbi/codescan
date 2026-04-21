// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package grammar

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// parseSource is a test helper that parses a Go source file and returns
// the comment group attached to its first top-level declaration, plus
// the FileSet used during parsing.
func parseSource(t *testing.T, src string) (*ast.CommentGroup, *token.FileSet) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(f.Decls) == 0 {
		t.Fatal("no decls in test source")
	}
	switch d := f.Decls[0].(type) {
	case *ast.GenDecl:
		if d.Doc != nil {
			return d.Doc, fset
		}
	case *ast.FuncDecl:
		if d.Doc != nil {
			return d.Doc, fset
		}
	}
	t.Fatal("decl has no doc comment")
	return nil, nil
}

func TestPreprocessNil(t *testing.T) {
	if got := Preprocess(nil, token.NewFileSet()); got != nil {
		t.Errorf("nil CommentGroup: want nil, got %v", got)
	}
	cg := &ast.CommentGroup{}
	if got := Preprocess(cg, nil); got != nil {
		t.Errorf("nil FileSet: want nil, got %v", got)
	}
}

func TestPreprocessSingleLineComment(t *testing.T) {
	cg, fset := parseSource(t, "package p\n\n// swagger:model Foo\ntype Foo struct{}\n")
	lines := Preprocess(cg, fset)
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d: %+v", len(lines), lines)
	}
	if lines[0].Text != "swagger:model Foo" {
		t.Errorf("text: got %q want %q", lines[0].Text, "swagger:model Foo")
	}
	if lines[0].Pos.Line != 3 {
		t.Errorf("line: got %d want 3", lines[0].Pos.Line)
	}
}

func TestPreprocessMultipleLineComments(t *testing.T) {
	src := `package p

// swagger:model Foo
// maximum: 10
// minimum: 0
type Foo int
`
	cg, fset := parseSource(t, src)
	lines := Preprocess(cg, fset)
	want := []string{"swagger:model Foo", "maximum: 10", "minimum: 0"}
	if len(lines) != len(want) {
		t.Fatalf("want %d lines, got %d", len(want), len(lines))
	}
	for i, w := range want {
		if lines[i].Text != w {
			t.Errorf("line %d text: got %q want %q", i, lines[i].Text, w)
		}
		if lines[i].Pos.Line != 3+i {
			t.Errorf("line %d: pos.Line = %d want %d", i, lines[i].Pos.Line, 3+i)
		}
	}
}

func TestPreprocessBlockComment(t *testing.T) {
	src := `package p

/*
 * swagger:model Foo
 * maximum: 10
 */
type Foo int
`
	cg, fset := parseSource(t, src)
	lines := Preprocess(cg, fset)
	// Expect 4 lines: empty first, two content, empty last.
	if len(lines) != 4 {
		t.Fatalf("want 4 lines, got %d: %+v", len(lines), lines)
	}
	if lines[1].Text != "swagger:model Foo" {
		t.Errorf("line 1: got %q want %q", lines[1].Text, "swagger:model Foo")
	}
	if lines[2].Text != "maximum: 10" {
		t.Errorf("line 2: got %q want %q", lines[2].Text, "maximum: 10")
	}
	// Positions should increment.
	if lines[1].Pos.Line != lines[0].Pos.Line+1 {
		t.Errorf("block-comment line positions must increment: %+v", lines)
	}
}

func TestPreprocessStripsMarkdownTablePipe(t *testing.T) {
	src := `package p

// | swagger:model Foo |
type Foo int
`
	cg, fset := parseSource(t, src)
	lines := Preprocess(cg, fset)
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(lines))
	}
	// Leading pipe stripped; content (including trailing pipe) preserved.
	if lines[0].Text != "swagger:model Foo |" {
		t.Errorf("got %q want %q", lines[0].Text, "swagger:model Foo |")
	}
}

func TestPreprocessPreservesEmbeddedWhitespace(t *testing.T) {
	src := `package p

//     indented content
type Foo int
`
	cg, fset := parseSource(t, src)
	lines := Preprocess(cg, fset)
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(lines))
	}
	// trimContentPrefix strips leading whitespace; embedded spaces
	// inside Text remain.
	if lines[0].Text != "indented content" {
		t.Errorf("got %q want %q", lines[0].Text, "indented content")
	}
}

func TestPreprocessMultiCommentGroup(t *testing.T) {
	// A comment group with multiple *ast.Comment entries separated by
	// only whitespace — Go groups them into a single CommentGroup.
	src := `package p

// first
// second
// third
type Foo int
`
	cg, fset := parseSource(t, src)
	if len(cg.List) < 2 {
		t.Fatalf("expected multi-entry CommentGroup, got %d", len(cg.List))
	}
	lines := Preprocess(cg, fset)
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d", len(lines))
	}
	want := []string{"first", "second", "third"}
	for i, w := range want {
		if lines[i].Text != w {
			t.Errorf("line %d: got %q want %q", i, lines[i].Text, w)
		}
	}
}
