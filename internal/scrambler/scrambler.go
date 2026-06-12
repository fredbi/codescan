// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

// Package scrambler minimizes and de-identifies a Go source tree while
// preserving exactly what codescan reads, so a maintainer can be handed a small,
// shareable bug repro that still reproduces the reported behavior.
//
// This is the L0 (Minimize) milestone: prune everything the scanner never reads
// — unreachable declarations, function bodies, now-unused imports — so the
// produced Swagger spec is byte-identical to the original. The package is a pure
// files-in/files-out AST transform: it never imports codescan; the fidelity
// oracle (re-scan and compare) is a test/consumer concern.
//
// Design: .claude/plans/anonymizer-repro-tool.md (§2 oracle, §3 contract, §4/§12
// transforms). Build plan: .claude/plans/scrambler-L0-build.md.
package scrambler

// Options configures a Minimize run.
type Options struct {
	// Dir is the module / work directory (= codescan Options.WorkDir).
	Dir string
	// Patterns are the package patterns to load, e.g. []string{"./..."}.
	Patterns []string
	// BuildTags is passed through to the loader as -tags.
	BuildTags string
	// Files is an optional overlay of in-memory source (absolute path → bytes).
	// When nil the loader reads from disk; the TUI issue-report mode feeds its
	// in-memory buffer here. See plan §12.2.
	Files map[string][]byte
}

// Stats reports what Minimize did, for surfacing in the UI.
type Stats struct {
	FilesIn       int
	FilesOut      int
	DeclsKept     int
	DeclsDropped  int
	BodiesEmptied int
}

// Result is the outcome of Minimize: the surviving source tree plus accounting.
type Result struct {
	// Files maps each surviving file's path to its reprinted bytes. Files pruned
	// entirely are absent (and listed in Dropped).
	Files map[string][]byte
	// Dropped lists files removed entirely (all their decls were unreachable).
	Dropped []string
	Stats   Stats
}

// Minimize loads the packages described by opts, prunes everything the scanner
// never reads — unreachable declarations, function bodies, now-unused imports —
// and returns the resulting (smaller) source tree.
//
// Invariant (verified by the oracle test): the spec codescan produces from the
// returned tree is byte-identical to the spec from the original.
func Minimize(opts Options) (*Result, error) {
	l, err := load(opts)
	if err != nil {
		return nil, err
	}

	roots := findRoots(l)
	c := computeClosure(l, roots)

	stats := &Stats{}
	prune(l, c, stats)

	res, err := emit(l, stats)
	if err != nil {
		return nil, err
	}
	res.Stats = *stats
	return res, nil
}
