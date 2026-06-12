// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package scrambler

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"

	"golang.org/x/tools/go/packages"
)

// prune applies the minimize transform to every in-scope file: drop decls not in
// the closure, empty kept function bodies, remove now-unused imports, and fix up
// the comment list so reprinting keeps exactly the surviving comments.
func prune(l *loaded, c *closure, stats *Stats) {
	for _, sf := range l.files {
		pruneFile(l.fset, sf, c, stats)
	}
}

func pruneFile(fset *token.FileSet, sf *sourceFile, c *closure, stats *Stats) {
	// Snapshot comment associations BEFORE mutation; re-derive file.Comments
	// AFTER all edits so dropped decls / emptied bodies / removed imports take
	// their comments with them and surviving annotations stay attached.
	origComments := sf.ast.Comments
	cmap := ast.NewCommentMap(fset, sf.ast, origComments)

	kept := make([]ast.Decl, 0, len(sf.ast.Decls))
	for _, decl := range sf.ast.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if obj := sf.pkg.TypesInfo.Defs[d.Name]; obj != nil && c.keep[obj] {
				// Bodies enclosing an annotation (a nested swagger:* type) are
				// load-bearing and kept verbatim; all others are emptied.
				if !bodyHasSwaggerAnnotation(d, origComments) && emptyBody(d) {
					stats.BodiesEmptied++
				}
				stats.DeclsKept++
				kept = append(kept, d)
			} else {
				stats.DeclsDropped++
			}
		case *ast.GenDecl:
			if d.Tok == token.IMPORT {
				kept = append(kept, d) // imports pruned below, once decls settle
				continue
			}
			d.Specs = filterSpecs(sf.pkg, d.Specs, c, stats)
			if len(d.Specs) > 0 {
				kept = append(kept, d)
			}
		default:
			stats.DeclsDropped++
		}
	}
	sf.ast.Decls = kept

	removeUnusedImports(sf.ast, sf.pkg)

	sf.ast.Comments = keepComments(cmap, sf.ast, origComments)
}

// keepComments re-derives the file's comment list after pruning: comments tied to
// surviving nodes (via the comment map) plus every annotation-bearing group that
// the map would otherwise drop — package-doc swagger:meta and floating
// swagger:route blocks have no prunable host but are load-bearing.
func keepComments(cmap ast.CommentMap, f *ast.File, orig []*ast.CommentGroup) []*ast.CommentGroup {
	out := cmap.Filter(f).Comments()
	seen := make(map[*ast.CommentGroup]bool, len(out))
	for _, g := range out {
		seen[g] = true
	}
	for _, g := range orig {
		if !seen[g] && hasSwaggerAnnotation(g) {
			out = append(out, g)
			seen[g] = true
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Pos() < out[j].Pos() })
	return out
}

func filterSpecs(pkg *packages.Package, specs []ast.Spec, c *closure, stats *Stats) []ast.Spec {
	out := specs[:0]
	for _, spec := range specs {
		if specKept(pkg, spec, c) {
			stats.DeclsKept++
			out = append(out, spec)
		} else {
			stats.DeclsDropped++
		}
	}
	return out
}

func specKept(pkg *packages.Package, spec ast.Spec, c *closure) bool {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		obj := pkg.TypesInfo.Defs[s.Name]
		return obj != nil && c.keep[obj]
	case *ast.ValueSpec:
		// const-block grouping in the closure guarantees all-or-nothing per
		// block, so any kept name means the whole spec is kept.
		for _, name := range s.Names {
			if obj := pkg.TypesInfo.Defs[name]; obj != nil && c.keep[obj] {
				return true
			}
		}
	}
	return false
}

// emptyBody replaces a kept function's body with a bare return, naming the
// results where needed (Go zero-initializes named results). Returns whether a
// body was actually emptied. Bodyless decls (interface methods, asm) are left
// untouched. Named vs unnamed results are the same function type, so interface
// satisfaction (and IsTextMarshaler recognition) is preserved.
func emptyBody(fd *ast.FuncDecl) bool {
	if fd.Body == nil {
		return false
	}
	if fd.Type.Results == nil || len(fd.Type.Results.List) == 0 {
		fd.Body = &ast.BlockStmt{}
		return true
	}
	nameResults(fd)
	fd.Body = &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{}}}
	return true
}

// nameResults gives every result a name (synthesizing v1, v2, … for unnamed or
// blank ones), avoiding collisions with receiver, type-param and param names.
func nameResults(fd *ast.FuncDecl) {
	used := map[string]bool{}
	note := func(fl *ast.FieldList) {
		if fl == nil {
			return
		}
		for _, f := range fl.List {
			for _, n := range f.Names {
				if n.Name != "_" {
					used[n.Name] = true
				}
			}
		}
	}
	note(fd.Recv)
	note(fd.Type.TypeParams)
	note(fd.Type.Params)
	note(fd.Type.Results)

	n := 0
	next := func() string {
		for {
			n++
			cand := fmt.Sprintf("v%d", n)
			if !used[cand] {
				used[cand] = true
				return cand
			}
		}
	}
	for _, f := range fd.Type.Results.List {
		if len(f.Names) == 0 {
			f.Names = []*ast.Ident{ast.NewIdent(next())}
			continue
		}
		for i, nm := range f.Names {
			if nm.Name == "_" {
				f.Names[i] = ast.NewIdent(next())
			}
		}
	}
}
