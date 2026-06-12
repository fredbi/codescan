// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package scrambler

import (
	"sort"
	"testing"

	"github.com/go-openapi/codescan/internal/scantest"
)

// namedBasic is a small one-package fixture with four annotated types
// (Email/Colour/Grade via swagger:strfmt|type|default, User via swagger:model).
func namedBasicOptions() Options {
	return Options{
		Dir:      scantest.FixturesDir(),
		Patterns: []string{"./enhancements/named-basic"},
	}
}

// TestLoad exercises T1: the loader resolves the fixture package and captures
// its files with source bytes.
func TestLoad(t *testing.T) {
	l, err := load(namedBasicOptions())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(l.files) == 0 {
		t.Fatal("no files loaded")
	}
	for _, sf := range l.files {
		if len(sf.src) == 0 {
			t.Errorf("file %q loaded with empty source", sf.path)
		}
		if sf.ast == nil {
			t.Errorf("file %q loaded without an AST", sf.path)
		}
	}
}

// TestFindRoots exercises T2: the annotated declarations are recognised as roots.
func TestFindRoots(t *testing.T) {
	l, err := load(namedBasicOptions())
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	got := map[string]bool{}
	for obj := range findRoots(l) {
		got[obj.Name()] = true
	}

	for _, want := range []string{"Email", "Colour", "Grade", "User"} {
		if !got[want] {
			t.Errorf("expected %q among roots; got %v", want, sortedKeys(got))
		}
	}
}

// TestMinimizeResultShape checks the Result accounting is internally consistent.
// named-basic is all annotated/reachable types in one file, so nothing is dropped.
func TestMinimizeResultShape(t *testing.T) {
	res, err := Minimize(namedBasicOptions())
	if err != nil {
		t.Fatalf("Minimize: %v", err)
	}
	if res.Stats.FilesIn == 0 {
		t.Fatal("expected at least one file in")
	}
	if len(res.Files) != res.Stats.FilesOut {
		t.Errorf("Files map size %d != FilesOut %d", len(res.Files), res.Stats.FilesOut)
	}
	if res.Stats.FilesIn != res.Stats.FilesOut+len(res.Dropped) {
		t.Errorf("FilesIn (%d) != FilesOut (%d) + Dropped (%d)",
			res.Stats.FilesIn, res.Stats.FilesOut, len(res.Dropped))
	}
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
