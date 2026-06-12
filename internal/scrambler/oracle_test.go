// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package scrambler_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-openapi/codescan"
	"github.com/go-openapi/codescan/internal/scantest"
	"github.com/go-openapi/codescan/internal/scrambler"
)

// TestOracle is the L0 correctness proof: for each fixture package set, the spec
// codescan produces from the minimized tree must be byte-identical to the spec
// from the original. Any divergence is a prune defect (we cut something
// load-bearing, mis-stripped a body, or removed a used import).
func TestOracle(t *testing.T) {
	cases := []struct {
		name     string
		patterns []string
	}{
		{"petstore", []string{"./goparsing/petstore/..."}},
		{"classification", []string{
			"./goparsing/classification",
			"./goparsing/classification/models",
			"./goparsing/classification/operations",
		}},
		{"named-basic", []string{"./enhancements/named-basic"}},
		{"embedded-types", []string{"./enhancements/embedded-types"}},
		{"enum-docs", []string{"./enhancements/enum-docs"}},
		{"defaults-examples", []string{"./enhancements/defaults-examples"}},
		{"interface-methods", []string{"./enhancements/interface-methods"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The oracle proves minimize FIDELITY, conditioned on the baseline
			// being scannable. If codescan can't scan the original, that's a
			// codescan issue, not the scrambler's — skip loudly rather than
			// report a false failure. (Defensive: every fixture here currently
			// scans clean; this guard keeps a future baseline regression from
			// masquerading as a scrambler defect.)
			orig, err := scanSpec(scantest.FixturesDir(), tc.patterns)
			if err != nil {
				t.Skipf("original scan failed — pre-existing codescan/grammar2 issue, not the scrambler: %v", err)
			}

			res, err := scrambler.Minimize(scrambler.Options{
				Dir:      scantest.FixturesDir(),
				Patterns: tc.patterns,
			})
			if err != nil {
				t.Fatalf("Minimize: %v", err)
			}

			tmp := materialize(t, res)
			got, err := scanSpec(tmp, tc.patterns)
			if err != nil {
				t.Fatalf("minimized tree no longer scans (a scrambler defect): %v", err)
			}

			if !bytes.Equal(orig, got) {
				t.Errorf("spec differs after minimize (%d→%d files, %d decls dropped, %d bodies emptied)\n%s",
					res.Stats.FilesIn, res.Stats.FilesOut, res.Stats.DeclsDropped, res.Stats.BodiesEmptied,
					firstDiff(orig, got))
			}
		})
	}
}

// scanSpec runs codescan over dir/patterns and returns the spec as indented JSON.
func scanSpec(dir string, patterns []string) ([]byte, error) {
	sw, err := codescan.Run(&codescan.Options{
		WorkDir:    dir,
		Packages:   patterns,
		ScanModels: true,
	})
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(sw, "", "  ")
}

// materialize copies the fixtures tree to a temp dir, applies the minimized
// files, removes the dropped ones, and returns the temp root.
func materialize(t *testing.T, res *scrambler.Result) string {
	t.Helper()
	root := scantest.FixturesDir()
	tmp := t.TempDir()
	if err := os.CopyFS(tmp, os.DirFS(root)); err != nil {
		t.Fatalf("copy fixtures: %v", err)
	}
	for abs, content := range res.Files {
		dst := translate(t, root, tmp, abs)
		if err := os.WriteFile(dst, content, 0o600); err != nil {
			t.Fatalf("write %s: %v", dst, err)
		}
	}
	for _, abs := range res.Dropped {
		if err := os.Remove(translate(t, root, tmp, abs)); err != nil {
			t.Fatalf("remove dropped %s: %v", abs, err)
		}
	}
	return tmp
}

func translate(t *testing.T, root, tmp, abs string) string {
	t.Helper()
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		t.Fatalf("rel(%s,%s): %v", root, abs, err)
	}
	return filepath.Join(tmp, rel)
}

// firstDiff returns a short, human-readable hint at the first byte difference.
func firstDiff(a, b []byte) string {
	n := min(len(a), len(b))
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	ctx := func(s []byte) string {
		lo := max(0, i-60)
		hi := min(len(s), i+60)
		return string(s[lo:hi])
	}
	return "first diff at byte " + strconv.Itoa(i) + "\n--- original ---\n" + ctx(a) + "\n--- minimized ---\n" + ctx(b)
}
