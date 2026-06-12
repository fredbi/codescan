// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package scrambler

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/packages"
)

// removeUnusedImports deletes import specs no longer referenced after pruning.
// Removal-only: it never reorders surviving imports and never adds new ones, so
// it stays a small toolchain-free AST pass (no goimports / `go` binary), which
// matters for the eventual WASM build. See plan §7.6.
//
// Accuracy matters in both directions — an unused import left behind is a compile
// error, and a used import removed is too — so we resolve via type info
// (TypesInfo.Uses → *types.PkgName) rather than name heuristics that could be
// fooled by a local identifier shadowing a package name.
func removeUnusedImports(f *ast.File, pkg *packages.Package) {
	used := referencedPackages(f, pkg)

	byPath := make(map[string]*types.Package, len(pkg.Types.Imports()))
	for _, ip := range pkg.Types.Imports() {
		byPath[ip.Path()] = ip
	}

	decls := f.Decls[:0]
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.IMPORT {
			decls = append(decls, d)
			continue
		}
		gd.Specs = filterImportSpecs(gd.Specs, byPath, used)
		if len(gd.Specs) > 0 {
			decls = append(decls, gd)
		}
	}
	f.Decls = decls
}

// referencedPackages collects the set of imported packages still referenced in
// the (already-pruned) file.
func referencedPackages(f *ast.File, pkg *packages.Package) map[*types.Package]bool {
	used := map[*types.Package]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		if pn, ok := pkg.TypesInfo.Uses[id].(*types.PkgName); ok {
			used[pn.Imported()] = true
		}
		return true
	})
	return used
}

func filterImportSpecs(specs []ast.Spec, byPath map[string]*types.Package, used map[*types.Package]bool) []ast.Spec {
	out := specs[:0]
	for _, spec := range specs {
		is, ok := spec.(*ast.ImportSpec)
		if !ok || keepImport(is, byPath, used) {
			out = append(out, spec)
		}
	}
	return out
}

func keepImport(is *ast.ImportSpec, byPath map[string]*types.Package, used map[*types.Package]bool) bool {
	// Blank (side-effect) and dot (unqualified) imports can't be tracked by
	// qualifier and are always kept.
	if is.Name != nil && (is.Name.Name == "_" || is.Name.Name == ".") {
		return true
	}
	path, err := strconv.Unquote(is.Path.Value)
	if err != nil {
		return true // malformed — leave it alone
	}
	p := byPath[path]
	if p == nil {
		return true // couldn't resolve — keep to avoid breaking the build
	}
	return used[p]
}
