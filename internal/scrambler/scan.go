// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package scrambler

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

// findRoots returns the set of package-level objects that carry a swagger
// annotation — the seeds for the reachability closure (T3). Detection is
// deliberately coarse (presence of a "swagger:" comment line); L0 needs only the
// roots, not a parse of the annotation. Over-inclusion is safe: it only keeps
// more, and the byte-identical-spec oracle still holds.
func findRoots(l *loaded) map[types.Object]bool {
	roots := map[types.Object]bool{}
	for _, sf := range l.files {
		for _, decl := range sf.ast.Decls {
			collectRootDecl(roots, sf, decl)
		}
	}
	return roots
}

func collectRootDecl(roots map[types.Object]bool, sf *sourceFile, decl ast.Decl) {
	pkg := sf.pkg
	switch d := decl.(type) {
	case *ast.FuncDecl:
		// A func is a root if its own doc is annotated, or if its body encloses
		// an annotation — e.g. a swagger:response type declared inside the body.
		if hasSwaggerAnnotation(d.Doc) || bodyHasSwaggerAnnotation(d, sf.ast.Comments) {
			addObject(roots, pkg, d.Name)
		}
	case *ast.GenDecl:
		// An annotation on the GenDecl applies to all its specs; otherwise each
		// spec is checked individually (its own doc or trailing comment).
		group := hasSwaggerAnnotation(d.Doc)
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				if group || hasSwaggerAnnotation(s.Doc) || hasSwaggerAnnotation(s.Comment) {
					addObject(roots, pkg, s.Name)
				}
			case *ast.ValueSpec:
				if group || hasSwaggerAnnotation(s.Doc) || hasSwaggerAnnotation(s.Comment) {
					for _, name := range s.Names {
						addObject(roots, pkg, name)
					}
				}
			}
		}
	}
}

func addObject(roots map[types.Object]bool, pkg *packages.Package, ident *ast.Ident) {
	if ident == nil {
		return
	}
	if obj := pkg.TypesInfo.Defs[ident]; obj != nil {
		roots[obj] = true
	}
}

// hasSwaggerAnnotation reports whether the comment group carries a swagger
// annotation anywhere in its text. We match "swagger:" anywhere on a line, not
// just as a line prefix: go-swagger allows leading text before the tag, e.g.
// "// GetPets swagger:route GET /pets pets listPets". We scan the raw comment
// text rather than CommentGroup.Text(), which strips directive-style lines
// (cf. memory project_lsp_diagnostics_target).
func hasSwaggerAnnotation(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	for _, c := range cg.List {
		if strings.Contains(c.Text, "swagger:") {
			return true
		}
	}
	return false
}

// bodyHasSwaggerAnnotation reports whether a function body encloses a comment
// carrying a swagger annotation. The parser leaves comments on declarations
// nested in a body floating in the file's comment list (not attached to the
// nested decl), so we match by position range. Such bodies must not be emptied.
func bodyHasSwaggerAnnotation(fd *ast.FuncDecl, comments []*ast.CommentGroup) bool {
	if fd.Body == nil {
		return false
	}
	lo, hi := fd.Body.Lbrace, fd.Body.Rbrace
	for _, cg := range comments {
		if cg.Pos() >= lo && cg.End() <= hi && hasSwaggerAnnotation(cg) {
			return true
		}
	}
	return false
}

// fileHasSwaggerAnnotation reports whether a file carries any annotation,
// including in its package doc (swagger:meta) or a floating comment
// (swagger:route) — annotations not attached to a prunable declaration.
func fileHasSwaggerAnnotation(f *ast.File) bool {
	if hasSwaggerAnnotation(f.Doc) {
		return true
	}
	return slices.ContainsFunc(f.Comments, hasSwaggerAnnotation)
}
