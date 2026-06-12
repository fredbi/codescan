// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package scrambler

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

// pkgLoadMode mirrors internal/scanner: we need full type info (NeedTypesInfo)
// for the reachability walk, syntax for the rewrite, and deps so external types
// resolve as leaves. Kept local to keep the package decoupled from the scanner.
const pkgLoadMode = packages.NeedName | packages.NeedFiles | packages.NeedImports |
	packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo

// sourceFile pairs a parsed in-scope file with its original bytes.
type sourceFile struct {
	path string
	pkg  *packages.Package
	ast  *ast.File
	src  []byte
}

// loaded holds the in-scope packages (the ones matching the patterns; their deps
// are leaves) plus their files. All packages share one FileSet.
type loaded struct {
	fset  *token.FileSet
	pkgs  []*packages.Package
	files []*sourceFile
}

// load loads the packages described by opts. It forces module mode (GOWORK=off)
// for the same reason codescan does: an ambient go.work silently switches `go
// list` to workspace mode, where relative "./..." patterns resolve to empty,
// errored packages. See memory project_gowork_scanning_gotcha.
func load(opts Options) (*loaded, error) {
	cfg := &packages.Config{
		Dir:     opts.Dir,
		Mode:    pkgLoadMode,
		Tests:   false,
		Env:     append(os.Environ(), "GOWORK=off"),
		Overlay: opts.Files,
	}
	if opts.BuildTags != "" {
		cfg.BuildFlags = []string{"-tags", opts.BuildTags}
	}

	pkgs, err := packages.Load(cfg, opts.Patterns...)
	if err != nil {
		return nil, fmt.Errorf("scrambler: loading packages: %w", err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("%w: no packages matched %v in %q", ErrScramble, opts.Patterns, opts.Dir)
	}
	if err := rootPackageErrors(pkgs); err != nil {
		return nil, err
	}

	l := &loaded{fset: pkgs[0].Fset, pkgs: pkgs}
	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			path := l.fset.Position(f.Pos()).Filename
			src, err := readSource(opts.Files, path)
			if err != nil {
				return nil, err
			}
			l.files = append(l.files, &sourceFile{path: path, pkg: pkg, ast: f, src: src})
		}
	}
	return l, nil
}

// readSource returns a file's bytes, preferring the overlay over disk.
func readSource(overlay map[string][]byte, path string) ([]byte, error) {
	if b, ok := overlay[path]; ok {
		return b, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("scrambler: reading %q: %w", path, err)
	}
	return b, nil
}

// rootPackageErrors fails fast if any in-scope package didn't type-check: the
// reachability walk relies on complete type info. We check only the root
// packages (not deps) to avoid noise from unrelated dependency diagnostics.
func rootPackageErrors(pkgs []*packages.Package) error {
	var msgs []string
	for _, p := range pkgs {
		for _, e := range p.Errors {
			msgs = append(msgs, e.Error())
		}
	}
	if len(msgs) > 0 {
		return fmt.Errorf("%w: package load errors:\n%s", ErrScramble, strings.Join(msgs, "\n"))
	}
	return nil
}
