// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package scrambler

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"strings"
)

// emit reprints every surviving file and assembles the Result. Files left with
// no declarations other than imports are dropped (their content was unreachable);
// test files are dropped too (defensive — the loader runs with Tests:false).
func emit(l *loaded, stats *Stats) (*Result, error) {
	res := &Result{Files: make(map[string][]byte, len(l.files))}
	for _, sf := range l.files {
		// Keep a file if it still declares something, or if it carries an
		// annotation with no prunable host (package-doc swagger:meta, floating
		// swagger:route). Otherwise it contributed nothing the scanner reads.
		if isTestFile(sf.path) || (!hasDeclContent(sf.ast) && !fileHasSwaggerAnnotation(sf.ast)) {
			res.Dropped = append(res.Dropped, sf.path)
			continue
		}
		var buf bytes.Buffer
		if err := format.Node(&buf, l.fset, sf.ast); err != nil {
			return nil, fmt.Errorf("%w: formatting %q: %w", ErrScramble, sf.path, err)
		}
		res.Files[sf.path] = buf.Bytes()
	}
	stats.FilesIn = len(l.files)
	stats.FilesOut = len(res.Files)
	return res, nil
}

// hasDeclContent reports whether the file has any declaration beyond imports.
func hasDeclContent(f *ast.File) bool {
	for _, d := range f.Decls {
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			continue
		}
		return true
	}
	return false
}

func isTestFile(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}
